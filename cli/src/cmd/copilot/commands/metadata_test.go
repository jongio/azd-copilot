package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ownedEnvVarPattern matches the environment variables azd-copilot defines for
// itself. AZD_SERVER and AZD_DEBUG are azd's, and AZURE_* values are forwarded
// from the project environment, so neither is owned here.
var ownedEnvVarPattern = regexp.MustCompile(`\b(AZD_COPILOT_[A-Z0-9_]+|COPILOT_CUSTOM_INSTRUCTIONS_DIRS|AZD_PROJECT_[A-Z0-9_]+|AZD_SERVICES|AZD_SERVICE_DETAILS|AZD_SUBSCRIPTION_[A-Z0-9_]+|AZD_TENANT_ID|AZD_USER|AZD_INFRA_[A-Z0-9_]+|AZD_HAS_BICEP)\b`)

// documentedEnvVars indexes the published list by name.
func documentedEnvVars(t *testing.T) map[string]string {
	t.Helper()

	documented := map[string]string{}
	for _, variable := range ExtensionEnvironmentVariables() {
		if _, duplicate := documented[variable.Name]; duplicate {
			t.Fatalf("environment variable %q is documented twice", variable.Name)
		}
		documented[variable.Name] = variable.Description
	}

	return documented
}

// TestEveryOwnedEnvVarIsDocumented walks the non-test source and fails if any
// environment variable this extension defines is missing from the published
// metadata. This is the guard that makes the published contract self
// maintaining: adding AZD_COPILOT_ANYTHING to the launcher without describing
// it here breaks the build rather than silently shipping an undocumented name.
func TestEveryOwnedEnvVarIsDocumented(t *testing.T) {
	documented := documentedEnvVars(t)

	root := filepath.Join("..", "..", "..")
	found := map[string][]string{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		for _, match := range ownedEnvVarPattern.FindAllString(string(content), -1) {
			found[match] = append(found[match], path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking source tree: %v", err)
	}

	if len(found) == 0 {
		t.Fatal("scanned the source tree and found no owned environment variables, so this guard is not actually checking anything")
	}

	for name, paths := range found {
		if _, ok := documented[name]; !ok {
			t.Errorf("environment variable %q is used in %v but is not described by ExtensionEnvironmentVariables", name, paths)
		}
	}
}

// TestDocumentedEnvVarsAreActuallyUsed is the other direction: a variable that
// no longer appears in the source is stale documentation, which is worse than
// none because a reader trusts it.
func TestDocumentedEnvVarsAreActuallyUsed(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	var combined strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if filepath.Base(path) == "metadata.go" {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		combined.Write(content)

		return nil
	})
	if err != nil {
		t.Fatalf("walking source tree: %v", err)
	}

	source := combined.String()
	for name := range documentedEnvVars(t) {
		if !strings.Contains(source, name) {
			t.Errorf("environment variable %q is documented but no longer appears anywhere in the source", name)
		}
	}
}

// TestEveryEnvVarHasADescriptionAndExample keeps the published document useful.
// A name with no description tells a reader nothing they could not get from
// grep.
func TestEveryEnvVarHasADescriptionAndExample(t *testing.T) {
	for _, variable := range ExtensionEnvironmentVariables() {
		if strings.TrimSpace(variable.Description) == "" {
			t.Errorf("environment variable %q has no description", variable.Name)
		}
		if strings.TrimSpace(variable.Example) == "" {
			t.Errorf("environment variable %q has no example", variable.Name)
		}
	}
}

// TestSetOnlyVarsSayTheyAreNotRead separates the one input variable from the
// fourteen output ones. An agent author reading this document needs to know
// which names it can set to influence behavior and which are only ever
// reported, and the descriptions are the only place that distinction lives.
func TestSetOnlyVarsSayTheyAreNotRead(t *testing.T) {
	readByExtension := map[string]bool{
		"AZD_COPILOT_DEBUG": true,
	}

	for _, variable := range ExtensionEnvironmentVariables() {
		saysNotRead := strings.Contains(variable.Description, "Not read by azd copilot")

		switch {
		case readByExtension[variable.Name] && saysNotRead:
			t.Errorf("%q is read by the extension but its description claims otherwise", variable.Name)
		case !readByExtension[variable.Name] && !saysNotRead:
			t.Errorf("%q is only set on the child process, so its description must say it is not read back", variable.Name)
		}
	}
}

// TestExtensionConfigurationOwnsNoConfigScopes pins the deliberate decision to
// publish no configuration schemas.
//
// azd-copilot persists no user configuration and defines no azure.yaml keys,
// so empty schemas would claim a surface that does not exist. If this
// extension ever starts writing configuration, this test fails and a real
// schema becomes owed, which is exactly when someone should be forced to
// write one.
func TestExtensionConfigurationOwnsNoConfigScopes(t *testing.T) {
	config := ExtensionConfiguration()

	if config.Global != nil {
		t.Error("Global schema is set, but azd-copilot persists no user configuration; add a schema deliberately or leave it nil")
	}
	if config.Project != nil {
		t.Error("Project schema is set, but azd-copilot defines no azure.yaml project keys")
	}
	if config.Service != nil {
		t.Error("Service schema is set, but azd-copilot defines no azure.yaml service keys")
	}
	if len(config.EnvironmentVariables) == 0 {
		t.Error("no environment variables published, so the configuration document says nothing at all")
	}
}

// TestNoUserConfigWrites backs the claim in ExtensionConfiguration's comment.
// The context command reads azd's user config for display; the moment
// something calls Set, the nil Global schema stops being honest.
func TestNoUserConfigWrites(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		if strings.Contains(string(content), "UserConfig().Set") {
			t.Errorf("%s writes azd user configuration, so ExtensionConfiguration must publish a Global schema describing what it writes", path)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walking source tree: %v", err)
	}
}

// TestMetadataCommandEmitsConfiguration runs the command end to end. The whole
// point of rebuilding it around GenerateExtensionMetadata rather than calling
// azdext.NewMetadataCommand is that the SDK helper drops Configuration on the
// floor, so the output is where that has to be proven.
func TestMetadataCommandEmitsConfiguration(t *testing.T) {
	root := &cobra.Command{Use: "copilot"}
	root.AddCommand(&cobra.Command{Use: "build", Short: "Build an app"})

	command := NewMetadataCommand(func() *cobra.Command { return root })

	var out bytes.Buffer
	command.SetOut(&out)
	command.SetArgs([]string{})

	if err := command.Execute(); err != nil {
		t.Fatalf("metadata command failed: %v", err)
	}

	var document struct {
		SchemaVersion string `json:"schemaVersion"`
		ID            string `json:"id"`
		Configuration *struct {
			Global               json.RawMessage `json:"global"`
			EnvironmentVariables []struct {
				Name string `json:"name"`
			} `json:"environmentVariables"`
		} `json:"configuration"`
	}

	if err := json.Unmarshal(out.Bytes(), &document); err != nil {
		t.Fatalf("metadata output is not valid json: %v\n%s", err, out.String())
	}

	if document.SchemaVersion != metadataSchemaVersion {
		t.Errorf("schemaVersion = %q, want %q", document.SchemaVersion, metadataSchemaVersion)
	}
	if document.ID != metadataExtensionID {
		t.Errorf("id = %q, want %q", document.ID, metadataExtensionID)
	}
	if document.Configuration == nil {
		t.Fatal("configuration is absent from the emitted metadata, which is the exact regression this command exists to prevent")
	}
	if document.Configuration.Global != nil {
		t.Error("global schema was emitted despite no configuration being owned")
	}

	got := len(document.Configuration.EnvironmentVariables)
	want := len(ExtensionEnvironmentVariables())
	if got != want {
		t.Errorf("emitted %d environment variables, want %d", got, want)
	}
}

// TestMetadataCommandIsHidden keeps it out of help output. azd calls it; users
// should not see it.
func TestMetadataCommandIsHidden(t *testing.T) {
	command := NewMetadataCommand(func() *cobra.Command { return &cobra.Command{Use: "copilot"} })

	if !command.Hidden {
		t.Error("metadata command is visible in help output")
	}
	if command.Use != "metadata" {
		t.Errorf("command name is %q, want \"metadata\"", command.Use)
	}
}

// TestExtensionIDMatchesManifest catches the drift that would make azd fail to
// associate this metadata with the installed extension.
func TestExtensionIDMatchesManifest(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "extension.yaml"))
	if err != nil {
		t.Fatalf("reading extension.yaml: %v", err)
	}

	if !strings.Contains(string(content), metadataExtensionID) {
		t.Errorf("extension.yaml does not contain the id %q used by the metadata command", metadataExtensionID)
	}
}

// failingWriter always fails, so the metadata command's write error branch gets
// exercised. Left uncovered it is unreachable in tests, which is how an
// unchecked write survived there in the first place.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestMetadataCommandReportsWriteFailure(t *testing.T) {
	command := NewMetadataCommand(func() *cobra.Command { return &cobra.Command{Use: "copilot"} })
	command.SetOut(failingWriter{})
	command.SetErr(failingWriter{})

	err := command.Execute()
	if err == nil {
		t.Fatal("Execute() = nil, want a write failure")
	}
	if !strings.Contains(err.Error(), "failed to write metadata") {
		t.Errorf("error = %v, want it to mention the metadata write", err)
	}
}

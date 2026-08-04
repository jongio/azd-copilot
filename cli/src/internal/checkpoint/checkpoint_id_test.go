// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateIDComponentsAcceptsKnownValues(t *testing.T) {
	for _, ct := range knownTypes {
		for _, p := range knownPhases {
			if err := validateIDComponents(ct, p); err != nil {
				t.Errorf("validateIDComponents(%q, %q) = %v, want nil", ct, p, err)
			}
		}
	}
}

// TestValidateIDComponentsAcceptsEmpty covers manual checkpoints, which
// legitimately carry no phase. Rejecting empty would break them.
func TestValidateIDComponentsAcceptsEmpty(t *testing.T) {
	if err := validateIDComponents("", ""); err != nil {
		t.Errorf("validateIDComponents(\"\", \"\") = %v, want nil", err)
	}
	if err := validateIDComponents(TypeManual, ""); err != nil {
		t.Errorf("validateIDComponents(TypeManual, \"\") = %v, want nil", err)
	}
}

func TestValidateIDComponentsRejectsTraversal(t *testing.T) {
	traversals := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config",
		"../outside",
		"..",
		"sub/dir",
		"sub\\dir",
	}

	for _, bad := range traversals {
		if err := validateIDComponents(TypePhase, Phase(bad)); err == nil {
			t.Errorf("validateIDComponents accepted phase %q, which reaches a filename", bad)
		}
		if err := validateIDComponents(CheckpointType(bad), PhaseSpec); err == nil {
			t.Errorf("validateIDComponents accepted type %q, which reaches a filename", bad)
		}
	}
}

func TestValidateIDComponentsRejectsUnknownValues(t *testing.T) {
	if err := validateIDComponents(TypePhase, Phase("not-a-phase")); err == nil {
		t.Error("validateIDComponents accepted an unknown phase")
	}
	if err := validateIDComponents(CheckpointType("not-a-type"), PhaseSpec); err == nil {
		t.Error("validateIDComponents accepted an unknown type")
	}
}

// TestSaveWithOptionsWritesNothingOnTraversal is the test that matters. The unit
// tests above prove the validator says no; this proves SaveWithOptions consults
// it before the index write, so a rejected phase leaves no trace on disk.
func TestSaveWithOptionsWritesNothingOnTraversal(t *testing.T) {
	root := t.TempDir()
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o750); err != nil {
		t.Fatalf("create work dir: %v", err)
	}
	t.Chdir(work)

	_, err := SaveWithOptions(SaveOptions{
		Type:        TypePhase,
		Phase:       Phase(filepath.Join("..", "..", "escaped")),
		Trigger:     TriggerManual,
		Description: "traversal attempt",
	})
	if err == nil {
		t.Fatal("SaveWithOptions accepted a traversal phase")
	}
	if !strings.Contains(err.Error(), "invalid checkpoint phase") {
		t.Errorf("error = %v, want an invalid phase rejection", err)
	}

	var written []string
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			written = append(written, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk %s: %v", root, walkErr)
	}
	if len(written) > 0 {
		t.Errorf("SaveWithOptions wrote %v despite rejecting the phase", written)
	}
}

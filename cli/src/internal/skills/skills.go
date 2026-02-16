package skills

import (
	"embed"

	"github.com/jongio/azd-copilot/cli/src/cmd/copilot/commands"
	"github.com/jongio/azd-core/copilotskills"
)

//go:embed azd-copilot/SKILL.md
var skillFS embed.FS

// InstallSkill installs the azd-copilot self-skill to ~/.copilot/skills/azd-copilot/
// This is separate from the 29 Azure skills in internal/assets/skills.go
func InstallSkill() error {
	return copilotskills.Install("azd-copilot", commands.Version, skillFS, "azd-copilot")
}

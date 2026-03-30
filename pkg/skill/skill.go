package skill

import (
	"cmp"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/yardenshoham/cobra-to-skills/pkg/discovery"
)

// Config holds configuration for skill generation.
type Config struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	AllowedTools  string
	Metadata      map[string]string
	Notes         []string
}

// GenerateDir generates the skill directory on the filesystem.
func GenerateDir(tree *discovery.CommandTree, dir string, config Config, logger *slog.Logger) error {
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write SKILL.md
	skillPath := filepath.Join(dir, "SKILL.md")
	logger.Info("writing SKILL.md", "path", skillPath)
	f, err := os.Create(skillPath)
	if err != nil {
		return fmt.Errorf("creating SKILL.md: %w", err)
	}
	defer f.Close()
	if err := GenerateSkill(tree, f, config); err != nil {
		return fmt.Errorf("generating SKILL.md: %w", err)
	}

	// Write reference files
	allCmds := discovery.AllCommands(tree)
	logger.Info("writing reference files", "count", len(allCmds))
	for _, cmd := range allCmds {
		refName := refFilename(cmd.CommandPath)
		refPath := filepath.Join(dir, "references", refName)
		logger.Debug("writing reference", "command", cmd.CommandPath, "path", refPath)
		rf, err := os.Create(refPath)
		if err != nil {
			return fmt.Errorf("creating reference file %s: %w", refName, err)
		}
		if err := GenerateReference(cmd, rf); err != nil {
			rf.Close()
			return fmt.Errorf("generating reference file %s: %w", refName, err)
		}
		rf.Close()
	}

	return nil
}

// GenerateSkill writes the SKILL.md content to w.
func GenerateSkill(tree *discovery.CommandTree, w io.Writer, config Config) error {
	name := cmp.Or(config.Name, tree.Name)
	description := cmp.Or(config.Description, tree.Long, tree.Short)

	// YAML frontmatter
	fmt.Fprintf(w, "---\n")
	fmt.Fprintf(w, "name: %s\n", name)
	fmt.Fprintf(w, "description: %q\n", description)
	if config.License != "" {
		fmt.Fprintf(w, "license: %s\n", config.License)
	}
	if config.Compatibility != "" {
		fmt.Fprintf(w, "compatibility: %q\n", config.Compatibility)
	}
	if config.AllowedTools != "" {
		fmt.Fprintf(w, "allowed-tools: %s\n", config.AllowedTools)
	}
	if len(config.Metadata) > 0 {
		fmt.Fprintf(w, "metadata:\n")
		for k, v := range config.Metadata {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
	}
	fmt.Fprintf(w, "---\n")

	// Body
	fmt.Fprintf(w, "\n# %s\n", name)

	// Long description
	if long := cmp.Or(tree.Long, tree.Short); long != "" {
		fmt.Fprintf(w, "\n%s\n", long)
	}

	// Notes
	if len(config.Notes) > 0 {
		fmt.Fprintf(w, "\n## Notes\n\n")
		for _, note := range config.Notes {
			fmt.Fprintf(w, "- %s\n", note)
		}
	}

	// Available Commands
	allCmds := discovery.AllCommands(tree)
	// Skip the root command itself
	nonRoot := allCmds[1:]
	if len(nonRoot) > 0 {
		fmt.Fprintf(w, "\n## Available Commands\n\n")
		for _, cmd := range nonRoot {
			refName := refFilename(cmd.CommandPath)
			fmt.Fprintf(w, "- [`%s`](references/%s) - %s\n", cmd.CommandPath, refName, cmp.Or(cmd.Short, cmd.Long))
		}
	}

	fmt.Fprintf(w, "\nSee [references/%s](references/%s) for root command flags.\n", refFilename(tree.CommandPath), refFilename(tree.CommandPath))
	fmt.Fprintf(w, "\nRun `%s --help` or `%s <command> --help` for full usage details.\n", name, name)

	return nil
}

// GenerateReference writes a reference file for a single command to w.
func GenerateReference(cmd *discovery.CommandTree, w io.Writer) error {
	// Heading
	fmt.Fprintf(w, "# %s\n", cmd.CommandPath)

	// Short description
	if cmd.Short != "" {
		fmt.Fprintf(w, "\n%s\n", cmd.Short)
	}

	// Long description (only if different from Short)
	if cmd.Long != "" && cmd.Long != cmd.Short {
		fmt.Fprintf(w, "\n%s\n", cmd.Long)
	}

	// Usage line (if runnable)
	if cmd.Runnable && cmd.UseLine != "" {
		fmt.Fprintf(w, "\n```\n%s\n```\n", cmd.UseLine)
	}

	// Examples
	if cmd.Example != "" {
		fmt.Fprintf(w, "\n## Examples\n\n```\n%s\n```\n", cmd.Example)
	}

	// Options (from Flags)
	if cmd.Flags != "" {
		fmt.Fprintf(w, "\n### Options\n\n```\n%s\n```\n", cmd.Flags)
	}

	// Options inherited from parent commands (from GlobalFlags)
	if cmd.GlobalFlags != "" {
		fmt.Fprintf(w, "\n### Options inherited from parent commands\n\n```\n%s\n```\n", cmd.GlobalFlags)
	}

	// Trailing newline
	fmt.Fprint(w, "\n")

	return nil
}

// refFilename returns the reference filename for a command path.
// e.g., "velero backup create" -> "velero_backup_create.md".
func refFilename(commandPath string) string {
	return strings.ReplaceAll(commandPath, " ", "_") + ".md"
}

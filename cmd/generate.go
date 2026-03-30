package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yardenshoham/cobra-to-skills/pkg/discovery"
	"github.com/yardenshoham/cobra-to-skills/pkg/skill"
)

func newGenerateCmd() *cobra.Command {
	var (
		output        string
		name          string
		description   string
		license       string
		compatibility string
		allowedTools  string
		metadata      []string
		notes         []string
	)

	cmd := &cobra.Command{
		Use:   "generate COBRA_BASED_CLI_PATH",
		Short: "Generate AI Agent skills from a Cobra CLI binary",
		Long: `Generate an agentskills.io-compatible skill directory from any Cobra-based CLI binary.

The command iteratively calls the target binary with --help to discover its full
command tree, then generates a SKILL.md and per-command reference files following
the Agent Skills specification for progressive disclosure.`,
		Example: `  # Generate a skill for a local binary
  cobra-to-skills generate ./my-cli --output ./skills/my-cli

  # Generate a skill for a binary on PATH
  cobra-to-skills generate kubectl --output ./skills/kubectl

  # Generate a skill with custom metadata
  cobra-to-skills generate velero --output ./skills --name velero \
    --description "Back up and restore Kubernetes cluster resources" \
    --license Apache-2.0`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryPath := args[0]
			logger, _ := cmd.Context().Value(loggerKey{}).(*slog.Logger)
			if logger == nil {
				logger = slog.New(slog.DiscardHandler)
			}

			logger.Info("starting discovery", "binary", binaryPath)
			tree, err := discovery.Discover(cmd.Context(), binaryPath, logger)
			if err != nil {
				return fmt.Errorf("discovering commands: %w", err)
			}

			config := skill.Config{
				Name:          name,
				Description:   description,
				License:       license,
				Compatibility: compatibility,
				AllowedTools:  allowedTools,
				Notes:         notes,
			}

			if len(metadata) > 0 {
				config.Metadata = make(map[string]string)
				for _, m := range metadata {
					key, value, ok := strings.Cut(m, "=")
					if !ok {
						return fmt.Errorf("invalid metadata format %q, expected key=value", m)
					}
					config.Metadata[key] = value
				}
			}

			if err := skill.GenerateDir(tree, output, config, logger); err != nil {
				return fmt.Errorf("generating skill: %w", err)
			}

			logger.Info("skill generation complete", "output", output)

			fmt.Fprintf(cmd.OutOrStdout(), "Skill generated at %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory for the generated skill")
	_ = cmd.MarkFlagRequired("output")
	cmd.Flags().StringVar(&name, "name", "", "Override skill name (default: binary name)")
	cmd.Flags().StringVar(&description, "description", "", "Override skill description (default: from root --help)")
	cmd.Flags().StringVar(&license, "license", "", "License identifier (e.g., Apache-2.0)")
	cmd.Flags().StringVar(&compatibility, "compatibility", "", "Compatibility requirements description")
	cmd.Flags().StringVar(&allowedTools, "allowed-tools", "", "Allowed tools specification")
	cmd.Flags().StringSliceVar(&metadata, "metadata", nil, "Metadata key=value pairs (can be repeated)")
	cmd.Flags().StringSliceVar(&notes, "notes", nil, "Usage notes (can be repeated)")

	return cmd
}

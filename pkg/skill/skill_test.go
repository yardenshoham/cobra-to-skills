package skill

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yardenshoham/cobra-to-skills/pkg/discovery"
	"github.com/yardenshoham/cobra-to-skills/pkg/parser"
)

func newTestTree() *discovery.CommandTree {
	return &discovery.CommandTree{
		Command: parser.Command{
			Name:        "mycli",
			CommandPath: "mycli",
			Short:       "A test CLI tool",
			Long:        "A test CLI tool that does many useful things.",
			UseLine:     "mycli [command]",
			Flags:       "  -h, --help   help for mycli",
		},
		Children: []*discovery.CommandTree{
			{
				Command: parser.Command{
					Name:        "sub1",
					CommandPath: "mycli sub1",
					Short:       "Do sub1 things",
					Long:        "Do sub1 things with more detail.",
					UseLine:     "mycli sub1 [flags]",
					Runnable:    true,
					Example:     "  mycli sub1 --name foo",
					Flags:       "  -n, --name string   the name",
					GlobalFlags: "  -v, --verbose   verbose output",
				},
			},
			{
				Command: parser.Command{
					Name:        "sub2",
					CommandPath: "mycli sub2",
					Short:       "Do sub2 things",
					Long:        "Do sub2 things",
					UseLine:     "mycli sub2 [command]",
					Flags:       "  -h, --help   help for sub2",
					GlobalFlags: "  -v, --verbose   verbose output",
				},
				Children: []*discovery.CommandTree{
					{
						Command: parser.Command{
							Name:        "leaf",
							CommandPath: "mycli sub2 leaf",
							Short:       "A leaf command",
							Long:        "A leaf command",
							UseLine:     "mycli sub2 leaf [flags]",
							Runnable:    true,
							Flags:       "  -h, --help   help for leaf",
							GlobalFlags: "  -v, --verbose   verbose output",
						},
					},
				},
			},
		},
	}
}

func TestGenerateSkill(t *testing.T) {
	t.Parallel()
	tree := newTestTree()
	var buf bytes.Buffer
	config := Config{
		Name:        "mycli",
		Description: "A test CLI tool",
	}
	err := GenerateSkill(tree, &buf, config)
	if err != nil {
		t.Fatalf("GenerateSkill failed: %v", err)
	}

	output := buf.String()

	// Check frontmatter
	if !strings.Contains(output, "name: mycli") {
		t.Error("missing name in frontmatter")
	}
	if !strings.Contains(output, `description: "A test CLI tool"`) {
		t.Error("missing description in frontmatter")
	}

	// Check heading
	if !strings.Contains(output, "# mycli") {
		t.Error("missing heading")
	}

	// Check Available Commands section
	if !strings.Contains(output, "## Available Commands") {
		t.Error("missing Available Commands section")
	}
	if !strings.Contains(output, "[`mycli sub1`](references/mycli_sub1.md) - Do sub1 things") {
		t.Error("missing sub1 in Available Commands")
	}
	if !strings.Contains(output, "[`mycli sub2`](references/mycli_sub2.md) - Do sub2 things") {
		t.Error("missing sub2 in Available Commands")
	}
	if !strings.Contains(output, "[`mycli sub2 leaf`](references/mycli_sub2_leaf.md) - A leaf command") {
		t.Error("missing sub2 leaf in Available Commands")
	}

	// Check footer
	if !strings.Contains(output, "See [references/mycli.md](references/mycli.md) for root command flags.") {
		t.Error("missing footer reference link")
	}
}

func TestGenerateReference(t *testing.T) {
	t.Parallel()

	t.Run("runnable command with examples", func(t *testing.T) {
		t.Parallel()
		cmd := &discovery.CommandTree{
			Command: parser.Command{
				Name:        "sub1",
				CommandPath: "mycli sub1",
				Short:       "Do sub1 things",
				Long:        "Do sub1 things with more detail.",
				UseLine:     "mycli sub1 [flags]",
				Runnable:    true,
				Example:     "  mycli sub1 --name foo",
				Flags:       "  -n, --name string   the name",
				GlobalFlags: "  -v, --verbose   verbose output",
			},
		}

		var buf bytes.Buffer
		err := GenerateReference(cmd, &buf)
		if err != nil {
			t.Fatalf("GenerateReference failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "# mycli sub1") {
			t.Error("missing heading")
		}
		if !strings.Contains(output, "Do sub1 things\n") {
			t.Error("missing short description")
		}
		if !strings.Contains(output, "Do sub1 things with more detail.") {
			t.Error("missing long description")
		}
		if !strings.Contains(output, "```\nmycli sub1 [flags]\n```") {
			t.Error("missing usage line")
		}
		if !strings.Contains(output, "## Examples") {
			t.Error("missing examples section")
		}
		if !strings.Contains(output, "### Options\n") {
			t.Error("missing options section")
		}
		if !strings.Contains(output, "### Options inherited from parent commands") {
			t.Error("missing inherited options section")
		}
	})

	t.Run("non-runnable command", func(t *testing.T) {
		t.Parallel()
		cmd := &discovery.CommandTree{
			Command: parser.Command{
				Name:        "sub2",
				CommandPath: "mycli sub2",
				Short:       "Do sub2 things",
				Long:        "Do sub2 things",
				UseLine:     "mycli sub2 [command]",
				Flags:       "  -h, --help   help for sub2",
			},
		}

		var buf bytes.Buffer
		err := GenerateReference(cmd, &buf)
		if err != nil {
			t.Fatalf("GenerateReference failed: %v", err)
		}

		output := buf.String()
		// Should not have usage code block (not runnable)
		if strings.Contains(output, "```\nmycli sub2 [command]\n```") {
			t.Error("non-runnable command should not have usage code block")
		}
	})
}

func TestRefFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"velero", "velero.md"},
		{"velero backup", "velero_backup.md"},
		{"velero backup create", "velero_backup_create.md"},
		{"velero backup-location create", "velero_backup-location_create.md"},
	}
	for _, tt := range tests {
		if got := refFilename(tt.path); got != tt.want {
			t.Errorf("refFilename(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

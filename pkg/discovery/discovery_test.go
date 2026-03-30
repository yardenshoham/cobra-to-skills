package discovery

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestDiscoverSelf(t *testing.T) {
	t.Parallel()

	// We need the cobra-to-skills binary. Try to find it.
	binary := "../../cobra-to-skills"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("cobra-to-skills binary not found, run 'go build' first")
	}

	tree, err := Discover(t.Context(), binary, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if tree.Name != "cobra-to-skills" {
		t.Errorf("root Name = %q, want %q", tree.Name, "cobra-to-skills")
	}

	// Should have at least version and completion (help is skipped)
	if len(tree.Children) < 2 {
		t.Errorf("expected at least 2 children, got %d", len(tree.Children))
	}

	names := map[string]bool{}
	for _, child := range tree.Children {
		names[child.Name] = true
	}
	if !names["version"] {
		t.Error("expected 'version' subcommand")
	}
	if !names["completion"] {
		t.Error("expected 'completion' subcommand")
	}
	if names["help"] {
		t.Error("'help' should be filtered out")
	}
}

func TestDiscoverWithMockExecutor(t *testing.T) {
	t.Parallel()

	helpOutputs := map[string]string{
		"--help": `My CLI tool

Usage:
  mycli [command]

Available Commands:
  sub1        Do sub1 things
  help        Help about any command

Flags:
  -h, --help   help for mycli

Use "mycli [command] --help" for more information about a command.`,
		"sub1 --help": `Do sub1 things

Usage:
  mycli sub1 [flags]

Flags:
  -n, --name string   the name

Global Flags:
  -v, --verbose   verbose output`,
	}

	executor := func(_ context.Context, binary string, args ...string) (string, error) {
		key := ""
		for _, a := range args {
			if key != "" {
				key += " "
			}
			key += a
		}
		if out, ok := helpOutputs[key]; ok {
			return out, nil
		}
		t.Fatalf("unexpected exec: %s %v", binary, args)
		return "", nil
	}

	tree, err := DiscoverWithExecutor(t.Context(), "mycli", executor, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if tree.Name != "mycli" {
		t.Errorf("root Name = %q, want %q", tree.Name, "mycli")
	}
	if tree.Short != "My CLI tool" {
		t.Errorf("root Short = %q", tree.Short)
	}

	// help should be filtered out, only sub1 remains
	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}
	child := tree.Children[0]
	if child.Name != "sub1" {
		t.Errorf("child Name = %q, want %q", child.Name, "sub1")
	}
	if child.Short != "Do sub1 things" {
		t.Errorf("child Short = %q, want %q", child.Short, "Do sub1 things")
	}
	if child.CommandPath != "mycli sub1" {
		t.Errorf("child CommandPath = %q, want %q", child.CommandPath, "mycli sub1")
	}
	if !child.Runnable {
		t.Error("sub1 should be runnable")
	}

	all := AllCommands(tree)
	if len(all) != 2 {
		t.Errorf("AllCommands returned %d, want 2", len(all))
	}
}

package discovery

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yardenshoham/cobra-to-skills/pkg/parser"
)

// CommandTree represents a discovered Cobra command with its children.
type CommandTree struct {
	parser.Command
	Children []*CommandTree
}

// skipCommands are Cobra built-in commands to skip during discovery.
var skipCommands = map[string]bool{
	"help": true,
}

// maxDepth is the maximum recursion depth for command tree discovery.
// Cobra CLIs rarely nest deeper than 5 levels; this guards against infinite loops.
const maxDepth = 15

// Executor runs a command and returns its output.
type Executor func(ctx context.Context, binary string, args ...string) (string, error)

// DefaultExecutor runs a binary with the given args and returns combined stdout+stderr.
func DefaultExecutor(ctx context.Context, binary string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	// Many Cobra apps exit 0 on --help, but some don't.
	// If we got output, use it regardless of exit code.
	out := stdout.String()
	if out == "" {
		out = stderr.String()
	}
	if out == "" && err != nil {
		return "", fmt.Errorf("failed to run %s %s: %w", binary, strings.Join(args, " "), err)
	}
	return out, nil
}

// Discover recursively discovers the full command tree of a Cobra-based binary.
func Discover(ctx context.Context, binary string, logger *slog.Logger) (*CommandTree, error) {
	return DiscoverWithExecutor(ctx, binary, DefaultExecutor, logger)
}

// DiscoverWithExecutor discovers the command tree using a custom executor.
func DiscoverWithExecutor(ctx context.Context, binary string, executor Executor, logger *slog.Logger) (*CommandTree, error) {
	name := filepath.Base(binary)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	logger.Info("discovering command tree", "binary", binary, "name", name)
	return discoverRecursive(ctx, binary, name, nil, "", executor, logger)
}

func discoverRecursive(ctx context.Context, binary, name string, args []string, parentShort string, executor Executor, logger *slog.Logger) (*CommandTree, error) {
	if len(args) > maxDepth {
		return nil, fmt.Errorf("maximum discovery depth (%d) exceeded at %s %s", maxDepth, binary, strings.Join(args, " "))
	}
	helpArgs := append(args, "--help")
	cmdPath := name
	if len(args) > 0 {
		cmdPath = name + " " + strings.Join(args, " ")
	}
	logger.Debug("executing help", "command", cmdPath, "args", helpArgs)
	output, err := executor(ctx, binary, helpArgs...)
	if err != nil {
		return nil, fmt.Errorf("discovering %s %s: %w", binary, strings.Join(args, " "), err)
	}
	logger.Debug("parsing help output", "command", cmdPath, "output_bytes", len(output))

	cmd := parser.Parse(output)

	// Set the command path based on the binary name and args
	if len(args) == 0 {
		cmd.CommandPath = name
		cmd.Name = name
	} else {
		cmd.CommandPath = name + " " + strings.Join(args, " ")
		cmd.Name = args[len(args)-1]
	}

	// If the parent provided a Short description from its Available Commands
	// listing, use that. The Long comes from the command's own --help.
	if parentShort != "" {
		cmd.Short = parentShort
	}

	tree := &CommandTree{Command: cmd}

	if len(cmd.SubcommandNames) > 0 {
		logger.Debug("found subcommands", "command", cmdPath, "count", len(cmd.SubcommandNames), "subcommands", cmd.SubcommandNames)
	}

	for _, subName := range cmd.SubcommandNames {
		if skipCommands[subName] {
			logger.Debug("skipping built-in command", "command", subName)
			continue
		}
		childShort := cmd.SubcommandShorts[subName]
		logger.Info("discovering subcommand", "parent", cmdPath, "subcommand", subName)
		child, err := discoverRecursive(ctx, binary, name, append(args, subName), childShort, executor, logger)
		if err != nil {
			return nil, err
		}
		tree.Children = append(tree.Children, child)
	}

	return tree, nil
}

// AllCommands returns a flat list of all commands in the tree (depth-first).
func AllCommands(tree *CommandTree) []*CommandTree {
	var result []*CommandTree
	var walk func(*CommandTree)
	walk = func(t *CommandTree) {
		result = append(result, t)
		for _, child := range t.Children {
			walk(child)
		}
	}
	walk(tree)
	return result
}

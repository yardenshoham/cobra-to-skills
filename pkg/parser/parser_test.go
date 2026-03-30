package parser //nolint:revive // matches the package under test

import (
	"os"
	"path/filepath"
	"testing"
)

func loadTestData(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", name, err)
	}
	return string(data)
}

func TestParseVeleroRoot(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "velero_root.txt")
	cmd := Parse(text)

	if cmd.Name != "velero" {
		t.Errorf("Name = %q, want %q", cmd.Name, "velero")
	}
	if cmd.CommandPath != "velero" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "velero")
	}
	if cmd.Runnable {
		t.Error("root command should not be runnable")
	}
	if cmd.Short != "Velero is a tool for managing disaster recovery, specifically for Kubernetes" {
		t.Errorf("Short = %q", cmd.Short)
	}
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
	if cmd.UseLine != "velero [command]" {
		t.Errorf("UseLine = %q, want %q", cmd.UseLine, "velero [command]")
	}
	if len(cmd.SubcommandNames) == 0 {
		t.Error("expected subcommand names")
	}
	// Should include both help and completion in raw parsing
	found := map[string]bool{}
	for _, name := range cmd.SubcommandNames {
		found[name] = true
	}
	for _, want := range []string{"backup", "restore", "schedule", "help", "completion"} {
		if !found[want] {
			t.Errorf("expected subcommand %q", want)
		}
	}
	if cmd.Flags == "" {
		t.Error("expected Flags to be set")
	}
	if cmd.GlobalFlags != "" {
		t.Error("root command should not have GlobalFlags (its flags are Flags)")
	}
	if cmd.Example != "" {
		t.Error("root command should not have Example")
	}
}

func TestParseVeleroBackup(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "velero_backup.txt")
	cmd := Parse(text)

	if cmd.Name != "backup" {
		t.Errorf("Name = %q, want %q", cmd.Name, "backup")
	}
	if cmd.CommandPath != "velero backup" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "velero backup")
	}
	if cmd.Short != "Work with backups" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Work with backups")
	}
	if cmd.Runnable {
		t.Error("backup group command should not be runnable")
	}
	if len(cmd.SubcommandNames) == 0 {
		t.Error("expected subcommand names")
	}
	if cmd.Flags == "" {
		t.Error("expected Flags to be set")
	}
	if cmd.GlobalFlags == "" {
		t.Error("expected GlobalFlags to be set")
	}
}

func TestParseVeleroBackupCreate(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "velero_backup_create.txt")
	cmd := Parse(text)

	if cmd.Name != "create" {
		t.Errorf("Name = %q, want %q", cmd.Name, "create")
	}
	if cmd.CommandPath != "velero backup create" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "velero backup create")
	}
	if cmd.Short != "Create a backup" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Create a backup")
	}
	if !cmd.Runnable {
		t.Error("backup create should be runnable")
	}
	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}
	if cmd.Flags == "" {
		t.Error("expected Flags to be set")
	}
	if cmd.GlobalFlags == "" {
		t.Error("expected GlobalFlags to be set")
	}
	if len(cmd.SubcommandNames) != 0 {
		t.Error("leaf command should not have subcommands")
	}
}

func TestParseVeleroCompletion(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "velero_completion.txt")
	cmd := Parse(text)

	if cmd.Name != "completion" {
		t.Errorf("Name = %q, want %q", cmd.Name, "completion")
	}
	if !cmd.Runnable {
		t.Error("completion should be runnable")
	}
	// Completion has a long multi-line description before Usage:
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestParseCobraToSkillsRoot(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "cobra-to-skills_root.txt")
	cmd := Parse(text)

	if cmd.Name != "cobra-to-skills" {
		t.Errorf("Name = %q, want %q", cmd.Name, "cobra-to-skills")
	}
	if cmd.Short != "Generate AI Agent skills from Cobra CLI applications" {
		t.Errorf("Short = %q", cmd.Short)
	}
	if cmd.Runnable {
		t.Error("root command should not be runnable")
	}
}

func TestParseKubectlRoot(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "kubectl_root.txt")
	cmd := Parse(text)

	if cmd.Name != "kubectl" {
		t.Errorf("Name = %q, want %q", cmd.Name, "kubectl")
	}
	if cmd.CommandPath != "kubectl" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "kubectl")
	}
	if cmd.Short != "kubectl controls the Kubernetes cluster manager." {
		t.Errorf("Short = %q", cmd.Short)
	}
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
	if len(cmd.SubcommandNames) == 0 {
		t.Fatal("expected subcommand names from categorized command groups")
	}
	// Verify subcommands from different categories are all discovered
	found := map[string]bool{}
	for _, name := range cmd.SubcommandNames {
		found[name] = true
	}
	for _, want := range []string{
		"create", "expose", "run", "set", // Basic Commands (Beginner)
		"get", "edit", "delete", "explain", // Basic Commands (Intermediate)
		"rollout", "scale", "autoscale", // Deploy Commands
		"certificate", "cluster-info", "top", // Cluster Management Commands
		"describe", "logs", "exec", "debug", // Troubleshooting
		"apply", "diff", "patch", // Advanced Commands
		"label", "annotate", "completion", // Settings Commands
		"api-resources", "config", "version", // Other Commands
	} {
		if !found[want] {
			t.Errorf("expected subcommand %q", want)
		}
	}
	// Should have 42 subcommands total across all categories
	if got := len(cmd.SubcommandNames); got != 42 {
		t.Errorf("got %d subcommands, want 42: %v", got, cmd.SubcommandNames)
	}
}

func TestParseNerdctlRoot(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "nerdctl_root.txt")
	cmd := Parse(text)

	if cmd.Name != "nerdctl" {
		t.Errorf("Name = %q, want %q", cmd.Name, "nerdctl")
	}
	if cmd.CommandPath != "nerdctl" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "nerdctl")
	}
	if len(cmd.SubcommandNames) == 0 {
		t.Fatal("expected subcommand names from Management commands + Commands groups")
	}
	found := map[string]bool{}
	for _, name := range cmd.SubcommandNames {
		found[name] = true
	}
	// Management commands
	for _, want := range []string{"apparmor", "builder", "container", "image", "ipfs", "namespace", "network", "system", "volume"} {
		if !found[want] {
			t.Errorf("expected management subcommand %q", want)
		}
	}
	// Commands (from the single-word "Commands:" header)
	for _, want := range []string{"build", "commit", "compose", "cp", "create", "exec", "run", "ps", "logs", "pull", "push", "images", "inspect", "kill", "rm", "stop", "start", "version"} {
		if !found[want] {
			t.Errorf("expected subcommand %q", want)
		}
	}
	if cmd.Flags == "" {
		t.Error("expected Flags to be set")
	}
}

func TestParseNerdctlContainer(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "nerdctl_container.txt")
	cmd := Parse(text)

	if cmd.Name != "container" {
		t.Errorf("Name = %q, want %q", cmd.Name, "container")
	}
	if cmd.CommandPath != "nerdctl container" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "nerdctl container")
	}
	if cmd.Short != "Manage containers" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Manage containers")
	}
	// "Commands:" is a single-word command group header — must be recognized
	if len(cmd.SubcommandNames) == 0 {
		t.Fatal("expected subcommand names from single-word Commands: header")
	}
	found := map[string]bool{}
	for _, name := range cmd.SubcommandNames {
		found[name] = true
	}
	for _, want := range []string{"commit", "cp", "create", "exec", "inspect", "kill", "logs", "ls", "pause", "prune", "rename", "restart", "rm", "run", "start", "stop", "unpause", "update", "wait"} {
		if !found[want] {
			t.Errorf("expected subcommand %q", want)
		}
	}
	// "port" should also be listed
	if !found["port"] {
		t.Error("expected subcommand \"port\"")
	}
	if got := len(cmd.SubcommandNames); got != 20 {
		t.Errorf("got %d subcommands, want 20: %v", got, cmd.SubcommandNames)
	}
}

func TestParseNerdctlImageDecrypt(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "nerdctl_image_decrypt.txt")
	cmd := Parse(text)

	if cmd.Name != "decrypt" {
		t.Errorf("Name = %q, want %q", cmd.Name, "decrypt")
	}
	if cmd.CommandPath != "nerdctl image decrypt" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "nerdctl image decrypt")
	}
	if !cmd.Runnable {
		t.Error("decrypt should be runnable")
	}
	if cmd.Flags == "" {
		t.Error("expected Flags to be set")
	}
	// Must NOT parse example lines like "openssl genrsa -out mykey.pem" as subcommands.
	// The "Example (encrypt):" and "Example (decrypt):" blocks contain indented lines
	// that look superficially like subcommand entries but are actually shell commands.
	if len(cmd.SubcommandNames) != 0 {
		t.Errorf("expected no subcommands, got %v", cmd.SubcommandNames)
	}
}

func TestParseKubectlGet(t *testing.T) {
	t.Parallel()
	text := loadTestData(t, "kubectl_get.txt")
	cmd := Parse(text)

	if cmd.Name != "get" {
		t.Errorf("Name = %q, want %q", cmd.Name, "get")
	}
	if cmd.CommandPath != "kubectl get" {
		t.Errorf("CommandPath = %q, want %q", cmd.CommandPath, "kubectl get")
	}
	if cmd.Short != "Display one or many resources." {
		t.Errorf("Short = %q, want %q", cmd.Short, "Display one or many resources.")
	}
	if !cmd.Runnable {
		t.Error("kubectl get should be runnable")
	}
	if cmd.Example == "" {
		t.Error("expected Example to be set")
	}
	// kubectl uses "Options:" instead of "Flags:" — should still be parsed
	if cmd.Flags == "" {
		t.Error("expected Flags to be set (from Options: section)")
	}
	if len(cmd.SubcommandNames) != 0 {
		t.Errorf("leaf command should not have subcommands, got %v", cmd.SubcommandNames)
	}
}

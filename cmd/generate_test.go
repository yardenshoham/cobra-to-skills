package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCmd(t *testing.T) {
	t.Parallel()

	// Build the binary first
	binary := filepath.Join("..", "cobra-to-skills")
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("cobra-to-skills binary not found, run 'go build' first")
	}

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "skills")

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"generate", binary, "--output", outputDir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check SKILL.md exists
	if _, err := os.Stat(filepath.Join(outputDir, "SKILL.md")); os.IsNotExist(err) {
		t.Error("SKILL.md not generated")
	}

	// Check references directory exists
	refs, err := os.ReadDir(filepath.Join(outputDir, "references"))
	if err != nil {
		t.Fatalf("Failed to read references directory: %v", err)
	}

	// Should have at least root + version + completion + generate
	if len(refs) < 3 {
		t.Errorf("Expected at least 3 reference files, got %d", len(refs))
	}
}

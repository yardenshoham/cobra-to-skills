package parser //nolint:revive // matches internal package name, not stdlib

import (
	"maps"
	"strings"
)

// Command represents a parsed Cobra command extracted from --help output.
type Command struct {
	Name             string
	CommandPath      string
	Short            string
	Long             string
	UseLine          string
	Example          string
	Flags            string
	GlobalFlags      string
	SubcommandNames  []string
	SubcommandShorts map[string]string
	Runnable         bool
}

// sectionHeaders are the recognized section headers in Cobra help output.
var sectionHeaders = []string{
	"Usage:",
	"Aliases:",
	"Available Commands:",
	"Examples:",
	"Flags:",
	"Options:",
	"Global Flags:",
	"Additional help topics:",
	`Use "`,
}

// Parse parses the text output of a Cobra command's --help into a Command struct.
func Parse(helpText string) Command {
	var cmd Command
	sections := splitSections(helpText)

	if desc, ok := sections["description"]; ok {
		desc = strings.TrimSpace(desc)
		cmd.Long = desc
		if before, _, found := strings.Cut(desc, "\n"); found {
			cmd.Short = before
		} else {
			cmd.Short = desc
		}
	}

	if usage, ok := sections["Usage:"]; ok {
		cmd.UseLine = firstNonEmptyLine(usage)
		cmd.Runnable = cmd.UseLine != "" && !strings.HasSuffix(cmd.UseLine, "[command]")
	}

	cmd.CommandPath, cmd.Name = extractCommandPath(cmd.UseLine)

	if avail, ok := sections["Available Commands:"]; ok {
		cmd.SubcommandNames, cmd.SubcommandShorts = parseSubcommands(avail)
	} else {
		cmd.SubcommandNames, cmd.SubcommandShorts = parseCommandGroups(helpText)
	}

	if ex, ok := sections["Examples:"]; ok {
		cmd.Example = trimTrailingEmptyLines(ex)
	}

	if flags, ok := sections["Flags:"]; ok {
		cmd.Flags = trimTrailingEmptyLines(flags)
	} else if flags, ok := sections["Options:"]; ok {
		cmd.Flags = trimTrailingEmptyLines(flags)
	}

	if gflags, ok := sections["Global Flags:"]; ok {
		cmd.GlobalFlags = trimTrailingEmptyLines(gflags)
	}

	return cmd
}

// firstNonEmptyLine returns the first non-empty, trimmed line from text.
func firstNonEmptyLine(text string) string {
	for line := range strings.SplitSeq(strings.TrimSpace(text), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// extractCommandPath parses a usage line like "velero backup create NAME [flags]"
// and returns the command path ("velero backup create") and the leaf name ("create").
// Words starting with [ or < and ALL_CAPS placeholders mark the end of the path.
func extractCommandPath(useLine string) (commandPath, name string) {
	parts := strings.Fields(useLine)
	if len(parts) == 0 {
		return "", ""
	}
	var path []string
	for _, p := range parts {
		if p[0] == '[' || p[0] == '<' || isAllUpper(p) {
			break
		}
		path = append(path, p)
	}
	if len(path) == 0 {
		return parts[0], parts[0]
	}
	return strings.Join(path, " "), path[len(path)-1]
}

// isAllUpper returns true if s is non-empty and consists entirely of uppercase
// ASCII letters (used to detect argument placeholders like NAME, FILE).
func isAllUpper(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return s != ""
}

// splitSections splits help text into named sections based on Cobra's section headers.
func splitSections(text string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(text, "\n")
	currentSection := "description"
	var buf []string

	for _, line := range lines {
		if header, inline := matchSectionHeader(line); header != "" {
			sections[currentSection] = strings.Join(buf, "\n")
			currentSection = header
			buf = nil
			if inline != "" {
				buf = append(buf, "  "+inline)
			}
		} else {
			buf = append(buf, line)
		}
	}
	if currentSection != "" {
		sections[currentSection] = strings.Join(buf, "\n")
	}
	return sections
}

// matchSectionHeader checks if a line is a recognized Cobra section header.
// Returns the header name and any inline content (e.g., "Usage: cmd [flags]"
// returns header="Usage:" and inlineContent="cmd [flags]").
func matchSectionHeader(line string) (header, inlineContent string) {
	trimmed := strings.TrimSpace(line)
	for _, h := range sectionHeaders {
		if h == `Use "` {
			if strings.HasPrefix(trimmed, `Use "`) {
				return h, ""
			}
			continue
		}
		if trimmed == h {
			return h, ""
		}
		if after, ok := strings.CutPrefix(trimmed, h+" "); ok {
			return h, strings.TrimSpace(after)
		}
	}
	return "", ""
}

// isCommandGroupHeader returns true if line looks like a command group header
// (e.g. "Deploy Commands:", "Management commands:") — a flush-left, short
// label ending with ":" that starts with a letter and isn't sentence-like.
func isCommandGroupHeader(line string) bool {
	if line == "" || line[0] == ' ' || line[0] == '\t' {
		return false
	}
	trimmed := strings.TrimSpace(line)
	if !strings.HasSuffix(trimmed, ":") {
		return false
	}
	if (line[0] < 'a' || line[0] > 'z') && (line[0] < 'A' || line[0] > 'Z') {
		return false
	}
	if strings.ContainsAny(trimmed, ",.;") {
		return false
	}
	wordCount := len(strings.Fields(strings.TrimSuffix(trimmed, ":")))
	return wordCount >= 1 && wordCount <= 5
}

// parseCommandGroups scans help text for categorized command group blocks
// (e.g. kubectl's "Deploy Commands:" followed by indented subcommand lines).
func parseCommandGroups(helpText string) ([]string, map[string]string) {
	var allNames []string
	allShorts := make(map[string]string)
	lines := strings.Split(helpText, "\n")

	for i := 0; i < len(lines); i++ {
		if h, _ := matchSectionHeader(lines[i]); h != "" {
			continue
		}
		if !isCommandGroupHeader(lines[i]) {
			continue
		}
		j := i + 1
		var block []string
		for j < len(lines) && (lines[j] == "" || lines[j][0] == ' ' || lines[j][0] == '\t') {
			block = append(block, lines[j])
			j++
		}
		names, shorts := parseSubcommands(strings.Join(block, "\n"))
		if len(names) > 0 {
			allNames = append(allNames, names...)
			maps.Copy(allShorts, shorts)
		}
		i = j - 1
	}
	return allNames, allShorts
}

// parseSubcommands parses indented two-column lines ("  name  description")
// from a command listing section.
func parseSubcommands(text string) ([]string, map[string]string) {
	var names []string
	shorts := make(map[string]string)
	for line := range strings.SplitSeq(text, "\n") {
		if name, desc := parseSubcommandLine(line); name != "" {
			names = append(names, name)
			shorts[name] = desc
		}
	}
	return names, shorts
}

// parseSubcommandLine extracts a subcommand name and description from a single
// Cobra-formatted line like "  name      description". Returns empty strings if
// the line isn't a valid subcommand entry. A valid entry must be indented, have
// a two-column layout (2+ space gap), a valid subcommand name, and no shell operators.
func parseSubcommandLine(line string) (name, description string) {
	if line == "" || (line[0] != ' ' && line[0] != '\t') {
		return "", ""
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.ContainsAny(trimmed, ">|<$") {
		return "", ""
	}
	// Cobra formats subcommand listings as two columns separated by 2+ spaces.
	n, rest, ok := strings.Cut(trimmed, "  ")
	if !ok || strings.TrimSpace(rest) == "" || !isValidSubcommandName(n) {
		return "", ""
	}
	return n, strings.TrimSpace(rest)
}

// isValidSubcommandName returns true if name is a valid Cobra subcommand:
// starts with a letter, contains only letters/digits/hyphens/underscores,
// and has at least one lowercase letter (rejecting ALL_CAPS placeholders).
func isValidSubcommandName(name string) bool {
	if name == "" {
		return false
	}
	if (name[0] < 'a' || name[0] > 'z') && (name[0] < 'A' || name[0] > 'Z') {
		return false
	}
	hasLower := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			// valid
		default:
			return false
		}
	}
	return hasLower
}

// trimTrailingEmptyLines removes leading and trailing empty lines from a text block.
func trimTrailingEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

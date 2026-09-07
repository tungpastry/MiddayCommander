package completion

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	execCacheMu sync.Mutex
	execPathEnv string
	execCache   []string
)

func CurrentWord(input string, position int) (int, int, string) {
	position = min(max(0, position), len(input))
	start := strings.LastIndexFunc(input[:position], func(r rune) bool {
		return strings.ContainsRune(" \t\n\r", r)
	})
	start++
	end := position
	for end < len(input) && !strings.ContainsRune(" \t\n\r", rune(input[end])) {
		end++
	}
	return start, end, input[start:end]
}

func CommonPrefix(values []string) string {
	if len(values) == 0 {
		return ""
	}
	prefix := values[0]
	for _, value := range values[1:] {
		for !strings.HasPrefix(value, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func CompletePathCandidates(prefix, baseDir string, directoriesOnly bool) []string {
	if prefix == "~" {
		return []string{"~/"}
	}
	rawDir, base := filepath.Split(prefix)
	if rawDir == "" {
		rawDir = "."
	}
	expandedDir := expandTilde(rawDir)
	scanDir := expandedDir
	if !filepath.IsAbs(scanDir) {
		scanDir = filepath.Join(baseDir, scanDir)
	}
	entries, err := os.ReadDir(scanDir)
	if err != nil {
		return nil
	}
	candidates := make([]string, 0, len(entries))
	for _, entry := range entries {
		if directoriesOnly && !entry.IsDir() {
			continue
		}
		if base != "" && !strings.HasPrefix(entry.Name(), base) {
			continue
		}
		candidate := filepath.Join(rawDir, entry.Name())
		if rawDir == "." {
			candidate = entry.Name()
		}
		if entry.IsDir() {
			candidate += string(filepath.Separator)
		}
		candidates = append(candidates, candidate)
	}
	sort.Strings(candidates)
	return candidates
}

func CompleteExecCandidates(prefix, baseDir string) []string {
	if strings.ContainsAny(prefix, `/\\`) {
		rawDir, base := filepath.Split(prefix)
		if rawDir == "" {
			rawDir = "."
		}
		scanDir := expandTilde(rawDir)
		if !filepath.IsAbs(scanDir) {
			scanDir = filepath.Join(baseDir, scanDir)
		}
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			return nil
		}
		var candidates []string
		for _, entry := range entries {
			if entry.IsDir() || (base != "" && !strings.HasPrefix(entry.Name(), base)) {
				continue
			}
			info, err := entry.Info()
			if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
				candidates = append(candidates, filepath.Join(rawDir, entry.Name()))
			}
		}
		sort.Strings(candidates)
		return candidates
	}

	pathEnv := os.Getenv("PATH")
	execCacheMu.Lock()
	if execCache == nil || pathEnv != execPathEnv {
		execPathEnv = pathEnv
		execCache = buildExecCache(pathEnv)
	}
	cached := append([]string(nil), execCache...)
	execCacheMu.Unlock()
	if prefix == "" {
		return cached
	}
	var candidates []string
	for _, candidate := range cached {
		if strings.HasPrefix(candidate, prefix) {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func FormatSuggestions(suggestions []string, width, maxLines int) []string {
	var lines []string
	line := " "
	for _, suggestion := range suggestions {
		display := filepath.Base(strings.TrimRight(suggestion, string(filepath.Separator)))
		separator := ""
		if line != " " {
			separator = "  "
		}
		if lipgloss.Width(line+separator+display) > width {
			lines = append(lines, pad(line, width))
			if len(lines) >= maxLines {
				return lines
			}
			line = " " + display
		} else {
			line += separator + display
		}
	}
	if line != " " && len(lines) < maxLines {
		lines = append(lines, pad(line, width))
	}
	return lines
}

func expandTilde(value string) string {
	if value != "~" && !strings.HasPrefix(value, "~/") {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return value
	}
	if value == "~" {
		return home
	}
	return filepath.Join(home, value[2:])
}

func buildExecCache(pathEnv string) []string {
	seen := make(map[string]struct{})
	var candidates []string
	for _, directory := range filepath.SplitList(pathEnv) {
		if directory == "" {
			directory = "."
		}
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if _, exists := seen[entry.Name()]; exists {
				continue
			}
			info, err := entry.Info()
			if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
				seen[entry.Name()] = struct{}{}
				candidates = append(candidates, entry.Name())
			}
		}
	}
	sort.Strings(candidates)
	return candidates
}

func pad(value string, width int) string {
	if lipgloss.Width(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-lipgloss.Width(value))
}

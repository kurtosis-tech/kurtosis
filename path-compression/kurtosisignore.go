package path_compression

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	kurtosisIgnoreFilename = ".kurtosisignore"
	commentPrefix          = "#"
)

// ignoreRules holds parsed ignore patterns from a .kurtosisignore file.
type ignoreRules struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	// The pattern to match against (e.g. ".git", "*.tmp", "static_files/erigon")
	pattern string
	// If true, pattern only matches directories
	dirOnly bool
	// If true, pattern contains a slash and should match against the full relative path
	hasSlash bool
}

// loadIgnoreRules reads a .kurtosisignore file from the given root directory.
// Returns nil if the file does not exist.
func loadIgnoreRules(rootPath string) *ignoreRules {
	ignoreFilePath := filepath.Join(rootPath, kurtosisIgnoreFilename)
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	rules := &ignoreRules{
		patterns: nil,
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, commentPrefix) {
			continue
		}

		p := ignorePattern{
			pattern:  "",
			dirOnly:  false,
			hasSlash: false,
		}

		// Trailing slash means directory-only match
		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		// If the pattern contains a slash (not just trailing), match against full relative path
		p.hasSlash = strings.Contains(line, "/")
		p.pattern = line

		rules.patterns = append(rules.patterns, p)
	}
	return rules
}

// shouldIgnore returns true if the given path should be ignored.
// relativePath is the path relative to the root (without leading separator).
// isDir indicates whether the path is a directory.
func (rules *ignoreRules) shouldIgnore(relativePath string, isDir bool) bool {
	if rules == nil {
		return false
	}

	// Normalize to forward slashes for matching
	relativePath = filepath.ToSlash(relativePath)
	baseName := filepath.Base(relativePath)

	for _, p := range rules.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		matchTarget := baseName
		if p.hasSlash {
			matchTarget = relativePath
		}

		matched, err := filepath.Match(p.pattern, matchTarget)
		if err != nil {
			// Invalid pattern, skip
			continue
		}
		if matched {
			return true
		}

		// For directory patterns without a slash, also check if the base name of any
		// path component matches (e.g. ".git" should match ".git/objects/pack")
		if !p.hasSlash {
			parts := strings.Split(relativePath, "/")
			for _, part := range parts {
				matched, err = filepath.Match(p.pattern, part)
				if err != nil {
					break
				}
				if matched {
					return true
				}
			}
		}
	}
	return false
}

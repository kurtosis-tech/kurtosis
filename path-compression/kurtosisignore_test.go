package path_compression

import (
	"encoding/hex"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKurtosisIgnore_ExcludesMatchingFiles(t *testing.T) {
	// Create a directory structure:
	// test-dir/
	// |-- .kurtosisignore  (contains ".git/" and "*.tmp")
	// |-- file_1.txt
	// |-- temp.tmp
	// |-- .git/
	// |   |-- config
	// |-- src/
	// |   |-- main.star
	dirPath, err := os.MkdirTemp("", "test-ignore-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	// Create .kurtosisignore
	ignoreContent := ".git/\n*.tmp\n"
	err = os.WriteFile(path.Join(dirPath, ".kurtosisignore"), []byte(ignoreContent), defaultPerm)
	require.NoError(t, err)

	// Create files
	f, err := os.Create(path.Join(dirPath, "file_1.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	f, err = os.Create(path.Join(dirPath, "temp.tmp"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	err = os.Mkdir(path.Join(dirPath, ".git"), defaultPerm)
	require.NoError(t, err)
	f, err = os.Create(path.Join(dirPath, ".git", "config"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	err = os.Mkdir(path.Join(dirPath, "src"), defaultPerm)
	require.NoError(t, err)
	f, err = os.Create(path.Join(dirPath, "src", "main.star"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Get the hash WITH .kurtosisignore
	hashWithIgnore, err := ComputeContentHash(dirPath)
	require.NoError(t, err)

	// Now create the same structure WITHOUT .git and *.tmp, no .kurtosisignore
	dirPath2, err := os.MkdirTemp("", "test-ignore-clean-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath2)

	f, err = os.Create(path.Join(dirPath2, ".kurtosisignore"))
	require.NoError(t, err)
	_, err = f.WriteString(".git/\n*.tmp\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	f, err = os.Create(path.Join(dirPath2, "file_1.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	err = os.Mkdir(path.Join(dirPath2, "src"), defaultPerm)
	require.NoError(t, err)
	f, err = os.Create(path.Join(dirPath2, "src", "main.star"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Without the ignored files, hash should be the same
	hashClean, err := ComputeContentHash(dirPath2)
	require.NoError(t, err)

	require.Equal(t, hex.EncodeToString(hashWithIgnore), hex.EncodeToString(hashClean),
		"Hash with .kurtosisignore excluding .git/ and *.tmp should match the clean directory")
}

func TestKurtosisIgnore_NoIgnoreFileNoEffect(t *testing.T) {
	dirPath, err := os.MkdirTemp("", "test-no-ignore-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	f, err := os.Create(path.Join(dirPath, "file_1.txt"))
	require.NoError(t, err)
	_, err = f.WriteString("content")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	hash1, err := ComputeContentHash(dirPath)
	require.NoError(t, err)

	hash2, err := ComputeContentHash(dirPath)
	require.NoError(t, err)

	require.Equal(t, hex.EncodeToString(hash1), hex.EncodeToString(hash2))
}

func TestKurtosisIgnore_CommentAndBlankLinesIgnored(t *testing.T) {
	dirPath, err := os.MkdirTemp("", "test-comments-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	ignoreContent := "# This is a comment\n\n*.tmp\n# Another comment\n"
	err = os.WriteFile(path.Join(dirPath, ".kurtosisignore"), []byte(ignoreContent), defaultPerm)
	require.NoError(t, err)

	f, err := os.Create(path.Join(dirPath, "keep.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	f, err = os.Create(path.Join(dirPath, "remove.tmp"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	rules := loadIgnoreRules(dirPath)
	require.NotNil(t, rules)
	require.Len(t, rules.patterns, 1)
	require.True(t, rules.shouldIgnore("remove.tmp", false))
	require.False(t, rules.shouldIgnore("keep.txt", false))
}

func TestKurtosisIgnore_DirOnlyPatterns(t *testing.T) {
	rules := &ignoreRules{
		patterns: []ignorePattern{
			{pattern: ".git", dirOnly: true, hasSlash: false},
		},
	}

	require.True(t, rules.shouldIgnore(".git", true))
	require.False(t, rules.shouldIgnore(".git", false))
	require.False(t, rules.shouldIgnore(".gitignore", false))
}

func TestKurtosisIgnore_PathPatterns(t *testing.T) {
	rules := &ignoreRules{
		patterns: []ignorePattern{
			{pattern: "static_files/erigon", dirOnly: false, hasSlash: true},
		},
	}

	require.True(t, rules.shouldIgnore("static_files/erigon", true))
	require.False(t, rules.shouldIgnore("erigon", true))
	require.False(t, rules.shouldIgnore("other/erigon", true))
}

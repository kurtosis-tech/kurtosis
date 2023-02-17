package git_package_content_provider

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/yaml_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	packagesDirRelPath    = "startosis-packages"
	packagesTmpDirRelPath = "tmp-startosis-packages"
)

func TestGitPackageProvider_SucceedsForValidPackage(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithExplicitMasterSet(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@master"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithBranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@test-branch"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Maybe!\"\n", contents)
}

func TestGitPackageProvider_FailsForInvalidBranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@non-existent-branch"
	_, err = provider.GetModuleContents(sampleStartosisModule)
	require.NotNil(t, err)
}

func TestGitPackageProvider_SucceedsForValidPackageWithTag(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@0.1.1"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithCommit(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@ec9062828e1a687a5db7dfa750f754f88119e4c0"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithCommitOnABranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@df88baf51caffbe7e8f66c0e54715f680f4482b2"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Maybe!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForNonStarlarkFile(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStarlarkPackage := "github.com/kurtosis-tech/eth2-package/static_files/prometheus-config/prometheus.yml.tmpl"
	contents, err := provider.GetModuleContents(sampleStarlarkPackage)
	require.Nil(t, err)
	require.NotEmpty(t, contents)
}

func TestGitPackageProvider_FailsForNonExistentPackage(t *testing.T) {
	oackageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(oackageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(oackageDir, packageTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = provider.GetModuleContents(nonExistentModulePath)
	require.NotNil(t, err)
}

func TestGetAbsolutePathOnDisk_WorksForPureDirectories(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	packagePath := "github.com/kurtosis-tech/datastore-army-package/src/helpers.star"
	pathOnDisk, err := provider.GetOnDiskAbsoluteFilePath(packagePath)

	require.Nil(t, err, "This test depends on your internet working and the kurtosis-tech/datastore-army-package existing")
	require.Equal(t, path.Join(packageDir, "kurtosis-tech", "datastore-army-package", "src/helpers.star"), pathOnDisk)
}

func Test_getPathToPackageRoot(t *testing.T) {
	githubUrlWithKurtosisPackageInSubfolder := "github.com/sample/sample-package/folder/subpackage"
	parsedGitUrl, err := parseGitURL(githubUrlWithKurtosisPackageInSubfolder)
	require.Nil(t, err, "Unexpected error occurred while parsing git url")
	actual := getPathToPackageRoot(parsedGitUrl)
	require.Equal(t, "sample/sample-package/folder/subpackage", actual)

	githubUrlWithRootKurtosisPackage := "github.com/sample/sample-package"
	parsedGitUrl, err = parseGitURL(githubUrlWithRootKurtosisPackage)
	require.Nil(t, err, "Unexpected error occurred while parsing git url")
	actual = getPathToPackageRoot(parsedGitUrl)
	require.Equal(t, "sample/sample-package", actual)
}

func Test_checkIfFileIsInAValidPackageInternal_somewhereInTheMiddle(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/data/packages/root/kurtosis.yml":                os.ErrNotExist,
			"/data/packages/root/subdir/kurtosis.yml":         os.ErrNotExist,
			"/data/packages/root/subdir/subdir1/kurtosis.yml": nil,
		}

		maybeError, found := filePathAndMockReturnMap[filePath]
		if !found {
			return nil, stacktrace.NewError("tried a path that was not accounted for %v", filePath)
		}

		return nil, maybeError
	}

	filePath := "/data/packages/root/subdir/subdir1/folder/some_file.txt"
	actual, err := checkIfFileIsInAValidPackageInternal(filePath, "/data/packages", mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, "/data/packages/root/subdir/subdir1/kurtosis.yml", actual)
}

func Test_checkIfFileIsInAValidPackageInternal_packageIsSameAsWhereTheFileIs(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/data/packages/root/kurtosis.yml":                os.ErrNotExist,
			"/data/packages/root/subdir/kurtosis.yml":         nil,
			"/data/packages/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		maybeError, found := filePathAndMockReturnMap[filePath]
		if !found {
			return nil, stacktrace.NewError("tried a path that was not accounted for %v", filePath)
		}

		return nil, maybeError
	}

	filePath := "/data/packages/root/subdir/some_file.txt"
	actual, err := checkIfFileIsInAValidPackageInternal(filePath, "/data/packages", mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, "/data/packages/root/subdir/kurtosis.yml", actual)
}

func Test_checkIfFileIsInAValidPackageInternal_fileNotFound(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/data/packages/root/kurtosis.yml":                os.ErrNotExist,
			"/data/packages/root/subdir/kurtosis.yml":         os.ErrNotExist,
			"/data/packages/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		maybeError, found := filePathAndMockReturnMap[filePath]
		if !found {
			return nil, stacktrace.NewError("tried a path that was not accounted for %v", filePath)
		}

		return nil, maybeError
	}

	filePath := "/data/packages/root/subdir/some_file.txt"
	actual, err := checkIfFileIsInAValidPackageInternal(filePath, "/data/packages", mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, filePathToKurtosisYamlNotFound, actual)
}

func Test_checkIfFileIsInAValidPackageInternal_unknownErrorOccurred(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/data/packages/root/kurtosis.yml":                os.ErrNotExist,
			"/data/packages/root/subdir/kurtosis.yml":         os.ErrClosed,
			"/data/packages/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		maybeError, found := filePathAndMockReturnMap[filePath]
		if !found {
			return nil, stacktrace.NewError("tried a path that was not accounted for %v", filePath)
		}
		return nil, maybeError
	}

	filePath := "/data/packages/root/subdir/some_file.txt"
	_, err := checkIfFileIsInAValidPackageInternal(filePath, "/data/packages", mockStatMethod)
	require.NotNil(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("An error occurred while locating %v in the path of '%v'", startosis_constants.KurtosisYamlName, filePath))
}

func Test_checkIfFileIsInAValidPackageInternal_prefixMismatchError(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		return nil, stacktrace.NewError("mock stat method should not be called")
	}

	filePath := "/packages/root/subdir/some_file.txt"
	_, err := checkIfFileIsInAValidPackageInternal(filePath, "/data/packages", mockStatMethod)
	require.NotNil(t, err)
	require.EqualError(t, err, fmt.Sprintf("Absolute path to file: %v must start with following prefix %v", filePath, "/data/packages"))
}

func Test_validateKurtosisPackage(t *testing.T) {
	type args struct {
		kurtosisYaml                    *yaml_parser.KurtosisYaml
		absPathToPackageWithKurtosisYml string
		packagesDir                     string
	}
	tests := []struct {
		name string
		args args
		want *startosis_errors.InterpretationError
	}{
		{
			name: "failure - mismatch package name and path (incorrect package name)",
			args: args{
				kurtosisYaml: &yaml_parser.KurtosisYaml{
					PackageName: "github.com/author/repo/packageIncorrect",
				},
				absPathToPackageWithKurtosisYml: "/root/folder/author/repo/package/kurtosis.yml",
				packagesDir:                     "/root/folder",
			},
			want: startosis_errors.NewInterpretationError("The package name in %v must match the location it is in. Package name is '%v' and it is found here:'%v'", startosis_constants.KurtosisYamlName, "github.com/author/repo/packageIncorrect", "github.com/author/repo/package"),
		},
		{
			name: "failure - mismatch package name and path (different location)",
			args: args{
				kurtosisYaml: &yaml_parser.KurtosisYaml{
					PackageName: "github.com/author/repo",
				},
				absPathToPackageWithKurtosisYml: "/root/folder/author/repo/subfolder/kurtosis.yml",
				packagesDir:                     "/root/folder",
			},
			want: startosis_errors.NewInterpretationError("The package name in %v must match the location it is in. Package name is '%v' and it is found here:'%v'", startosis_constants.KurtosisYamlName, "github.com/author/repo", "github.com/author/repo/subfolder"),
		},
		{
			name: "failure - contains a trailing '/'",
			args: args{
				kurtosisYaml: &yaml_parser.KurtosisYaml{
					PackageName: "github.com/author/repo/subfolder/",
				},
				absPathToPackageWithKurtosisYml: "/root/folder/author/repo/subfolder/kurtosis.yml",
				packagesDir:                     "/root/folder",
			},
			want: startosis_errors.NewInterpretationError("Error occurred while validating package name, found '%v' and expected '%v' without trailing '%v'", "github.com/author/repo/subfolder/", "github.com/author/repo/subfolder", string(os.PathSeparator)),
		},
		{
			name: "success - kurtosis.yml found in repo folder",
			args: args{
				kurtosisYaml: &yaml_parser.KurtosisYaml{
					PackageName: "github.com/author/repo",
				},
				absPathToPackageWithKurtosisYml: "/root/folder/author/repo/kurtosis.yml",
				packagesDir:                     "/root/folder",
			},
			want: nil,
		},
		{
			name: "success - kurtosis.yml found in sub folder folder",
			args: args{
				kurtosisYaml: &yaml_parser.KurtosisYaml{
					PackageName: "github.com/author/repo/subfolder",
				},
				absPathToPackageWithKurtosisYml: "/root/folder/author/repo/subfolder/kurtosis.yml",
				packagesDir:                     "/root/folder",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKurtosisPackage(tt.args.kurtosisYaml, tt.args.absPathToPackageWithKurtosisYml, tt.args.packagesDir)
			if tt.want == nil {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tt.want.Error())
			}
		})
	}
}

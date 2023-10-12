package kurtosis_package

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	isExecutablePackage    = true
	isNotExecutablePackage = false
)

func TestInitializeKurtosisPackage_Success(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", "package-initialize-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	// check kurtosis.yml file creation
	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}
	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.Equal(t, packageName, kurtosisYamlObj.PackageName)

	// check main.star file creation
	expectedMainStarFilepath := path.Join(packageDirpath, "main.star")
	fileBytes, err = os.ReadFile(expectedMainStarFilepath)
	require.NoError(t, err)
	require.Equal(t, mainStarFileContentStr, string(fileBytes))
}

func TestInitializeKurtosisPackage_IsNotExecutablePackageSuccess(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", "package-initialize-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isNotExecutablePackage)
	require.NoError(t, err)

	// check kurtosis.yml file creation
	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}
	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.Equal(t, packageName, kurtosisYamlObj.PackageName)

	// check main.star file was not created
	expectedMainStarFilepath := path.Join(packageDirpath, "main.star")
	fileBytes, err = os.ReadFile(expectedMainStarFilepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestInitializeKurtosisPackage_FailsIfKurtosisYmlAlreadyExist(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", "package-initialize-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	secondPackageName := "github.com/org/second-package"
	err = InitializeKurtosisPackage(packageDirpath, secondPackageName, isExecutablePackage)
	require.Error(t, err)
	expectedErrorMsgPortion := "'kurtosis.yml' already exist on this path"
	require.Contains(t, stacktrace.RootCause(err).Error(), expectedErrorMsgPortion)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}

	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)

	require.Equal(t, packageName, kurtosisYamlObj.PackageName)
}

func TestInitializeKurtosisPackage_FailsIfMainStarFileAlreadyExist(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", "package-initialize-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")

	err = os.Remove(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	secondPackageName := "github.com/org/second-package"
	err = InitializeKurtosisPackage(packageDirpath, secondPackageName, isExecutablePackage)
	require.Error(t, err)
	expectedErrorMsgPortion := "'main.star' already exist on this path"
	require.Contains(t, stacktrace.RootCause(err).Error(), expectedErrorMsgPortion)

	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}

	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)

	require.Equal(t, secondPackageName, kurtosisYamlObj.PackageName)
}

func TestInitializeKurtosisPackage_InvalidPackageNameError(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", "package-initialize-test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "my-rul/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid GitHub URL")
}

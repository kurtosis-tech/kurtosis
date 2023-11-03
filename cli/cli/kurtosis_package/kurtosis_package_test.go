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
	isExecutablePackage             = true
	isNotExecutablePackage          = false
	myTestPackage                   = "github.com/org/my-package"
	mySecondTestPackage             = "github.com/org/second-package"
	packageInitializeTestDirPattern = "package-initialize-test-dir-*"
)

func TestInitializeKurtosisPackage_Success(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := myTestPackage
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	// check kurtosis.yml file creation
	expectedKurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{
		PackageName:           "",
		PackageDescription:    "",
		PackageReplaceOptions: map[string]string{},
	}
	err = yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.NoError(t, err)
	require.Equal(t, packageName, kurtosisYamlObj.PackageName)

	// check main.star file creation
	expectedMainStarFilepath := path.Join(packageDirpath, mainStarFilename)
	fileBytes, err = os.ReadFile(expectedMainStarFilepath)
	require.NoError(t, err)
	require.Equal(t, mainStarFileContentStr, string(fileBytes))
}

func TestInitializeKurtosisPackage_IsNotExecutablePackageSuccess(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := myTestPackage
	err = InitializeKurtosisPackage(packageDirpath, packageName, isNotExecutablePackage)
	require.NoError(t, err)

	// check kurtosis.yml file creation
	expectedKurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{
		PackageName:           "",
		PackageDescription:    "",
		PackageReplaceOptions: map[string]string{},
	}
	err = yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.NoError(t, err)
	require.Equal(t, packageName, kurtosisYamlObj.PackageName)

	// check main.star file was not created
	expectedMainStarFilepath := path.Join(packageDirpath, mainStarFilename)
	fileBytes, err = os.ReadFile(expectedMainStarFilepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Nil(t, fileBytes)
}

func TestInitializeKurtosisPackage_FailsIfKurtosisYmlAlreadyExist(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := myTestPackage
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	secondPackageName := mySecondTestPackage
	err = InitializeKurtosisPackage(packageDirpath, secondPackageName, isExecutablePackage)
	require.Error(t, err)
	expectedErrorMsgPortion := "'kurtosis.yml' already exist on this path"
	require.Contains(t, stacktrace.RootCause(err).Error(), expectedErrorMsgPortion)

	kurtosisYamlObj := &enclaves.KurtosisYaml{
		PackageName:           "",
		PackageDescription:    "",
		PackageReplaceOptions: map[string]string{},
	}

	err = yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.NoError(t, err)

	require.Equal(t, packageName, kurtosisYamlObj.PackageName)
}

func TestInitializeKurtosisPackage_FailsIfMainStarFileAlreadyExist(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	// there is a main.star file inside the folder before initializing the package
	mainStarFilepath := path.Join(packageDirpath, mainStarFilename)

	_, err = os.Create(mainStarFilepath)
	require.NoError(t, err)

	packageName := myTestPackage
	err = InitializeKurtosisPackage(packageDirpath, packageName, isNotExecutablePackage)
	require.NoError(t, err)

	// check kurtosis.yml file creation
	expectedKurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{
		PackageName:           "",
		PackageDescription:    "",
		PackageReplaceOptions: map[string]string{},
	}
	err = yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.NoError(t, err)
	require.Equal(t, packageName, kurtosisYamlObj.PackageName)

}

func TestInitializeKurtosisPackage_InvalidPackageNameError(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := "my-url/org/my-package"
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid GitHub URL")
}

func TestInitializeKurtosisPackage_MakeScriptPackageSuccess(t *testing.T) {
	packageDirpath, err := os.MkdirTemp("", packageInitializeTestDirPattern)
	require.NoError(t, err)
	defer os.RemoveAll(packageDirpath)

	packageName := myTestPackage
	err = InitializeKurtosisPackage(packageDirpath, packageName, isExecutablePackage)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, kurtosisYmlFilename)
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	secondPackageName := mySecondTestPackage
	err = InitializeKurtosisPackage(packageDirpath, secondPackageName, isExecutablePackage)
	require.Error(t, err)
	expectedErrorMsgPortion := "'kurtosis.yml' already exist on this path"
	require.Contains(t, stacktrace.RootCause(err).Error(), expectedErrorMsgPortion)

	kurtosisYamlObj := &enclaves.KurtosisYaml{
		PackageName:           "",
		PackageDescription:    "",
		PackageReplaceOptions: map[string]string{},
	}

	err = yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)
	require.NoError(t, err)

	require.Equal(t, packageName, kurtosisYamlObj.PackageName)
}

package kurtosis_yaml

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func TestInitializeKurtosisPackage_Success(t *testing.T) {
	packageDirpath := os.TempDir()
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err := initializeKurtosisPackage(packageDirpath, packageName)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}

	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)

	require.Equal(t, packageName, kurtosisYamlObj.PackageName)
}

func TestInitializeKurtosisPackage_FailsIfKurtosisYmlAlreadyExist(t *testing.T) {
	packageDirpath := os.TempDir()
	defer os.RemoveAll(packageDirpath)

	packageName := "github.com/org/my-package"
	err := initializeKurtosisPackage(packageDirpath, packageName)
	require.NoError(t, err)

	expectedKurtosisYamlFilepath := path.Join(packageDirpath, "kurtosis.yml")
	fileBytes, err := os.ReadFile(expectedKurtosisYamlFilepath)
	require.NoError(t, err)

	secondPackageName := "github.com/org/second-package"
	err = initializeKurtosisPackage(packageDirpath, secondPackageName)
	require.Error(t, err)
	expectedErrorMsgPortion := "'kurtosis.yml' already exist on this path"
	require.Contains(t, stacktrace.RootCause(err).Error(), expectedErrorMsgPortion)

	kurtosisYamlObj := &enclaves.KurtosisYaml{}

	yaml.UnmarshalStrict(fileBytes, kurtosisYamlObj)

	require.Equal(t, packageName, kurtosisYamlObj.PackageName)
}

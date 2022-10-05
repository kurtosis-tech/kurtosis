package git_module_manager

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	dirPermission = 0755
)

func TestStartosisInterpreter_GitModuleManagerSucceedsForExistentModule(t *testing.T) {
	moduleDir := "/tmp/startosis-modules/"
	err := os.Mkdir(moduleDir, dirPermission)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir := "/tmp/tmp-startosis-modules/"
	err = os.Mkdir(moduleTmpDir, dirPermission)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	gitModuleManager := NewGitModuleManager(moduleDir, moduleTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := gitModuleManager.GetModule(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestStartosisInterpreter_GitModuleManagerFailsForNonExistentModule(t *testing.T) {
	moduleDir := "/tmp/startosis-modules/"
	err := os.Mkdir(moduleDir, dirPermission)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir := "/tmp/tmp-startosis-modules/"
	err = os.Mkdir(moduleTmpDir, dirPermission)
	require.Nil(t, err)
	os.RemoveAll(moduleTmpDir)

	gitModuleManager := NewGitModuleManager(moduleDir, moduleTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = gitModuleManager.GetModule(nonExistentModulePath)
	require.NotNil(t, nonExistentModulePath)
}



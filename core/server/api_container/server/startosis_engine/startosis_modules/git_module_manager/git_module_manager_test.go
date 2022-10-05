package git_module_manager

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
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

	gitModuleManager := git_module_manager.NewGitModuleManager(moduleDir, moduleTmpDir)

	interpreter := NewStartosisInterpreter(testServiceNetwork, gitModuleManager)
	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	script := `
load("` + sampleStartosisModule + `", "a")
print("Hello " + a)
`
	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError, "This test requires you to be connected to GitHub, ignore if you are offline")

	expectedOutput := "Hello World!\n"
	assert.Equal(t, expectedOutput, string(scriptOutput))
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

	gitModuleManager := git_module_manager.NewGitModuleManager(moduleDir, moduleTmpDir)

	interpreter := NewStartosisInterpreter(testServiceNetwork, gitModuleManager)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"
	script := `
load("` + nonExistentModulePath + `", "b")
print(b)
`
	_, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		fmt.Sprintf("Evaluation error: cannot load %v: An error occurred while fetching contents of the module '%v'", nonExistentModulePath, nonExistentModulePath),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 1)),
		},
	)
	assert.Equal(t, expectedError, interpretationError)
}



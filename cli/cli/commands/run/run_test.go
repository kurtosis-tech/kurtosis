package run

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/stretchr/testify/require"
	"testing"
)

const testScriptArg = "."

var testCtx context.Context = nil
var testParsedFlags *flags.ParsedFlags = nil

func TestValidateArgsJson_valid(t *testing.T) {
	inputArgs := `{"hello": "world!"}`
	parsedArgs, err := args.ParseArgsForValidation(StarlarkRunCmd.Args, []string{testScriptArg, inputArgs})
	require.Nil(t, err)
	require.NotNil(t, parsedArgs)

	err = validatePackageArgs(testCtx, testParsedFlags, parsedArgs)
	require.Nil(t, err)
}

func TestValidateArgsYaml_valid(t *testing.T) {
	inputArgs := `hello: world`
	parsedArgs, err := args.ParseArgsForValidation(StarlarkRunCmd.Args, []string{testScriptArg, inputArgs})
	require.Nil(t, err)
	require.NotNil(t, parsedArgs)

	err = validatePackageArgs(testCtx, testParsedFlags, parsedArgs)
	require.Nil(t, err)
}

func TestValidateArgs_invalid(t *testing.T) {
	inputArgs := `"hello" - "world!"` // missing { }
	parsedArgs, err := args.ParseArgsForValidation(StarlarkRunCmd.Args, []string{testScriptArg, inputArgs})
	require.Nil(t, err)
	require.NotNil(t, parsedArgs)

	err = validatePackageArgs(testCtx, testParsedFlags, parsedArgs)
	require.NotNil(t, err)
}

func TestIsHttpUrl_ValidHTTP(t *testing.T) {
	fileUrl := "http://www.mysite.com/myfile.json"

	isHttpUrlResult := isHttpUrl(fileUrl)

	require.True(t, isHttpUrlResult)
}

func TestIsHttpUrl_ValidHTTPS(t *testing.T) {
	fileUrl := "https://www.mysite.com/myfile.json"

	isHttpUrlResult := isHttpUrl(fileUrl)

	require.True(t, isHttpUrlResult)
}

func TestIsHttpUrl_NoValidBecauseIsAbsoluteFilepath(t *testing.T) {
	fileUrl := "/my-folder/myfile.json"

	isHttpUrlResult := isHttpUrl(fileUrl)

	require.False(t, isHttpUrlResult)
}

func TestIsHttpUrl_NoValidBecauseIsRelativeFilepath(t *testing.T) {
	fileUrl := "../my-folder/myfile.json"

	isHttpUrlResult := isHttpUrl(fileUrl)

	require.False(t, isHttpUrlResult)
}

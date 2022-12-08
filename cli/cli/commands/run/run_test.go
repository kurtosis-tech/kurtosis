package run

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/stretchr/testify/require"
	"testing"
)

const testScriptArg = "."

func TestValidateArgs_valid(t *testing.T) {
	inputArgs := `{"hello": "world!"}`
	args, err := args.ParseArgsForValidation(StarlarkRunCmd.Args, []string{testScriptArg, inputArgs})
	require.Nil(t, err)
	require.NotNil(t, args)
}

func TestValidateArgs_invalid(t *testing.T) {
	inputArgs := `"hello": "world!"` // missing { }
	args, err := args.ParseArgsForValidation(StarlarkRunCmd.Args, []string{testScriptArg, inputArgs})
	require.Nil(t, err)
	require.NotNil(t, args)

}

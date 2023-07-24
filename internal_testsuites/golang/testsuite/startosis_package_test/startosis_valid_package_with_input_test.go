package startosis_package_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	validPackageWithInputRelPath = "../../../starlark/valid-kurtosis-package-with-input"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_ValidPackageWithInput() {
	ctx := context.Background()
	params := `{"greetings": "bonjour!"}`
	runResult, err := suite.RunPackageWithParams(ctx, validPackageWithInputRelPath, params)

	t := suite.T()
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `bonjour!
Hello World!
{
	"message": "Hello World!"
}
`
	require.Equal(t, expectedScriptOutput, string(runResult.RunOutput))
	require.Len(t, runResult.Instructions, 2)
	logrus.Info("Successfully ran Startosis module")
}

func (suite *StartosisPackageTestSuite) TestStartosisPackage_ValidPackageWithInput_MissingKeyInParams() {
	ctx := context.Background()
	params := `{"hello": "world"}` // expecting key 'greetings' here
	runResult, _ := suite.RunPackageWithParams(ctx, validPackageWithInputRelPath, params)

	t := suite.T()
	require.NotNil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), "Evaluation error: key \"greetings\" not in dict")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Empty(t, string(runResult.RunOutput))
	require.Empty(t, runResult.Instructions)
	logrus.Info("Successfully ran Startosis module")
}

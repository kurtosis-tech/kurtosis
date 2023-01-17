package builtin_argument

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	builtinName = "test_builtin"

	serviceIdArgName = "service_id"
	serviceId        = starlark.String("datastore-server")

	shouldStartArgName = "should_start"
	shouldStart        = starlark.Bool(true)
)

func TestParseArguments_SingleRequiredArgument_FromArgs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{
		serviceId,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 1)
	require.Equal(t, serviceId, values[0])
}

func TestParseArguments_SingleRequiredArgument_FromArgs_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{
		starlark.MakeInt(10), // argument definition expects a starlark.String here
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "The following argument could not be parse because their type did not match the expected: [service_id]")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromArgs_NoValueProvided(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{} // empty but service_id is a mandatory arg
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test_builtin: missing argument for service_id")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromArgs_TooManyArgs(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{
		serviceId,
		starlark.MakeInt(10), // unexpected argument
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test_builtin: got 2 arguments, want at most 1")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceIdArgName),
			serviceId,
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 1)
	require.Equal(t, serviceId, values[0])
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceIdArgName),
			starlark.MakeInt(10), // argument definition expects a starlark.String here
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "The following argument could not be parse because their type did not match the expected: [service_id]")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_TooManyArgs(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceId()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceIdArgName),
			serviceId,
		},
		[]starlark.Value{
			starlark.String("max_memory"), // unexpected argument here
			starlark.MakeInt(1024),
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), `test_builtin: unexpected keyword argument "max_memory"`)
	require.Empty(t, values)
}

func TestParseArguments_ArgumentWithOptional_FromArgsOnly_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	args := starlark.Tuple{
		serviceId,
		shouldStart,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceId, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromKwargsOnly_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceIdArgName),
			serviceId,
		},
		[]starlark.Value{
			starlark.String(shouldStartArgName),
			shouldStart,
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceId, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgsAndKwargs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	args := starlark.Tuple{
		serviceId,
	}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(shouldStartArgName),
			shouldStart,
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceId, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgsNoOptional_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	args := starlark.Tuple{
		serviceId,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceId, values[0])
	require.Equal(t, starlark.Bool(false), values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgs_FailureTypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	args := starlark.Tuple{
		shouldStart,
		serviceId,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "The following argument could not be parse because their type did not match the expected: [service_id should_start]")
	require.Nil(t, values)
}

func TestExtractArgumentValue_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceIdValue starlark.String
	err := argumentValuesSet.ExtractArgumentValue(serviceIdArgName, &serviceIdValue)
	require.Nil(t, err)
	require.Equal(t, serviceId, serviceIdValue)
}

func TestExtractArgumentValue_Failure_ArgumentNotFound(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceIdValue starlark.Int // expecting a string here
	err := argumentValuesSet.ExtractArgumentValue("unknown_argument", &serviceIdValue)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Argument 'unknown_argument' could not be found in schema")
}

func TestExtractArgumentValue_Failure_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceIdValue starlark.Int // expecting a string here
	err := argumentValuesSet.ExtractArgumentValue(serviceIdArgName, &serviceIdValue)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Unable to extract value for argument 'service_id'. Types were not assignable (got 'starlark.Int', expecting 'starlark.String')")
}

func TestExtractArgumentValueStatic_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	value, err := ExtractArgumentValue[starlark.String](argumentValuesSet, serviceIdArgName)
	require.Nil(t, err)
	require.Equal(t, serviceId, value)
}

func TestExtractArgumentValueStatic_Failure_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	_, err := ExtractArgumentValue[starlark.Int](argumentValuesSet, serviceIdArgName)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Unable to extract value for argument 'service_id'. Types were not assignable (got 'starlark.Int', expecting 'starlark.String')")
}

func TestString(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceIdAndShouldStart()
	values := []starlark.Value{
		serviceId,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	expectedString := fmt.Sprintf(`(service_id=%s, should_start=%s)`, serviceId.String(), shouldStart.String())
	require.Equal(t, expectedString, argumentValuesSet.String())
}

// Private test helpers only below \\
func getArgumentDefinitionsWithServiceId() []*BuiltinArgument {
	return []*BuiltinArgument{
		{
			Name:              serviceIdArgName,
			IsOptional:        false,
			ZeroValueProvider: ZeroValueProvider[starlark.String],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				panic("should not be called")
			},
		},
	}
}

func getArgumentDefinitionsWithServiceIdAndShouldStart() []*BuiltinArgument {
	return []*BuiltinArgument{
		{
			Name:              serviceIdArgName,
			IsOptional:        false,
			ZeroValueProvider: ZeroValueProvider[starlark.String],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				panic("should not be called")
			},
		},
		{
			Name:              shouldStartArgName,
			IsOptional:        true,
			ZeroValueProvider: ZeroValueProvider[starlark.Bool],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				panic("should not be called")
			},
		},
	}
}

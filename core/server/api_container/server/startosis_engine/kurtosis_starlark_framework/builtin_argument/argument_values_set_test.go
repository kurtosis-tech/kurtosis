package builtin_argument

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"reflect"
	"testing"
)

const (
	builtinName = "test_builtin"

	serviceNameArgName = "service_name"
	serviceName        = starlark.String("datastore-server")

	shouldStartArgName = "should_start"
	shouldStart        = starlark.Bool(true)
)

func TestParseArguments_SingleRequiredArgument_FromArgs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{
		serviceName,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 1)
	require.Equal(t, serviceName, values[0])
}

func TestParseArguments_SingleRequiredArgument_FromArgs_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{
		starlark.MakeInt(10), // argument definition expects a starlark.String here
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "the following argument(s) could not be parsed or did not pass validation:")
	require.Contains(t, err.Error(), "{\"service_name\":\"For 'service_name', expected 'starlark.String', got 'starlark.Int'\"}")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromArgs_NoValueProvided(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{} // empty but service_name is a mandatory arg
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test_builtin: missing argument for service_name")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromArgs_TooManyArgs(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{
		serviceName,
		starlark.MakeInt(10), // unexpected argument
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "test_builtin: got 2 arguments, want at most 1")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceNameArgName),
			serviceName,
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 1)
	require.Equal(t, serviceName, values[0])
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceNameArgName),
			starlark.MakeInt(10), // argument definition expects a starlark.String here
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "the following argument(s) could not be parsed or did not pass validation:")
	require.Contains(t, err.Error(), "{\"service_name\":\"For 'service_name', expected 'starlark.String', got 'starlark.Int'\"}")
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_FailsValidation(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceNameArgName),
			starlark.String(""), // empty string for service_name will fail validation
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "the following argument(s) could not be parsed or did not pass validation:")
	require.Contains(t, err.Error(), `{"service_name":"'service_name' should not be empty"}`)
	require.Empty(t, values)
}

func TestParseArguments_SingleRequiredArgument_FromKwargs_TooManyArgs(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceName()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceNameArgName),
			serviceName,
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
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	args := starlark.Tuple{
		serviceName,
		shouldStart,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceName, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromKwargsOnly_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	args := starlark.Tuple{}
	kwargs := []starlark.Tuple{
		[]starlark.Value{
			starlark.String(serviceNameArgName),
			serviceName,
		},
		[]starlark.Value{
			starlark.String(shouldStartArgName),
			shouldStart,
		},
	}

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceName, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgsAndKwargs_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	args := starlark.Tuple{
		serviceName,
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
	require.Equal(t, serviceName, values[0])
	require.Equal(t, shouldStart, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgsNoOptional_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	args := starlark.Tuple{
		serviceName,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.Nil(t, err)
	require.Len(t, values, 2)
	require.Equal(t, serviceName, values[0])
	require.Equal(t, nil, values[1])
}

func TestParseArguments_ArgumentWithOptional_FromArgs_FailureTypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	args := starlark.Tuple{
		shouldStart, // serviceName should come first
		serviceName,
	}
	var kwargs []starlark.Tuple

	values, err := parseArguments(argumentDefinitions, builtinName, args, kwargs)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "the following argument(s) could not be parsed or did not pass validation:")
	require.Contains(t, err.Error(), "the argument 'service_name' could not be parsed because their type ('starlark.Bool') did not match the expected ('starlark.String')")
	require.Contains(t, err.Error(), "the argument 'should_start' could not be parsed because their type ('starlark.String') did not match the expected ('starlark.Bool')")
	require.Nil(t, values)
}

func TestIsSet_AllArgumentsSet(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	require.True(t, argumentValuesSet.IsSet(serviceNameArgName))
	require.True(t, argumentValuesSet.IsSet(shouldStartArgName))
}

func TestIsSet_MissingOneArgument(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		nil,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	require.True(t, argumentValuesSet.IsSet(serviceNameArgName))
	require.False(t, argumentValuesSet.IsSet(shouldStartArgName))
}

func TestExtractArgumentValue_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceNameValue starlark.String
	err := argumentValuesSet.ExtractArgumentValue(serviceNameArgName, &serviceNameValue)
	require.Nil(t, err)
	require.Equal(t, serviceName, serviceNameValue)
}

func TestExtractArgumentValue_Failure_ArgumentNotFound(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceNameValue starlark.Int // expecting a string here
	err := argumentValuesSet.ExtractArgumentValue("unknown_argument", &serviceNameValue)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "argument 'unknown_argument' could not be found in schema")
}

func TestExtractArgumentValue_Failure_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	var serviceNameValue starlark.Int // expecting a string here
	err := argumentValuesSet.ExtractArgumentValue(serviceNameArgName, &serviceNameValue)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unable to extract value for argument 'service_name'. Types were not assignable (got 'starlark.Int', expecting 'starlark.String')")
}

func TestExtractArgumentValueStatic_Success(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	value, err := ExtractArgumentValue[starlark.String](argumentValuesSet, serviceNameArgName)
	require.Nil(t, err)
	require.Equal(t, serviceName, value)
}

func TestExtractArgumentValueStatic_Failure_TypeMismatch(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	_, err := ExtractArgumentValue[starlark.Int](argumentValuesSet, serviceNameArgName)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unable to extract value for argument 'service_name'. Types were not assignable (got 'starlark.Int', expecting 'starlark.String')")
}

func TestString(t *testing.T) {
	argumentDefinitions := getArgumentDefinitionsWithServiceNameAndShouldStart()
	values := []starlark.Value{
		serviceName,
		shouldStart,
	}
	argumentValuesSet := NewArgumentValuesSet(argumentDefinitions, values)
	expectedString := fmt.Sprintf(`(service_name=%s, should_start=%s)`, serviceName.String(), shouldStart.String())
	require.Equal(t, expectedString, argumentValuesSet.String())
}

// --- Private test helpers only below --- \\
func getArgumentDefinitionsWithServiceName() []*BuiltinArgument {
	return []*BuiltinArgument{
		{
			Name:              serviceNameArgName,
			IsOptional:        false,
			ZeroValueProvider: ZeroValueProvider[starlark.String],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				valueStr, ok := value.(starlark.String)
				if !ok {
					return startosis_errors.NewInterpretationError("For '%s', expected 'starlark.String', got '%s'", serviceNameArgName, reflect.TypeOf(value))
				}
				if len(valueStr.GoString()) == 0 {
					return startosis_errors.NewInterpretationError("'%s' should not be empty", serviceNameArgName)
				}
				return nil
			},
		},
	}
}

func getArgumentDefinitionsWithServiceNameAndShouldStart() []*BuiltinArgument {
	return []*BuiltinArgument{
		{
			Name:              serviceNameArgName,
			IsOptional:        false,
			ZeroValueProvider: ZeroValueProvider[starlark.String],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				return nil
			},
		},
		{
			Name:              shouldStartArgName,
			IsOptional:        true,
			ZeroValueProvider: ZeroValueProvider[starlark.Bool],
			Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
				return nil
			},
		},
	}
}

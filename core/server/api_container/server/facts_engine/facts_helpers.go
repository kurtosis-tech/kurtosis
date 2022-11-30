package facts_engine

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
)

func StringifyFactValue(factValue *kurtosis_core_rpc_api_bindings.FactValue) (string, error) {
	switch factValue.GetFactValue().(type) {
	case *kurtosis_core_rpc_api_bindings.FactValue_StringValue:
		return factValue.GetStringValue(), nil
	}
	return "", stacktrace.NewError("Fact value cannot be stringified (type was: %d)", reflect.TypeOf(factValue.GetFactValue()))
}

package recipe

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	GetHttpRecipeTypeName = "GetHttpRequestRecipe"

	getMethod        = "GET"
	emptyRequestBody = ""
	noContentType    = ""
)

func NewGetHttpRequestRecipeType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: GetHttpRecipeTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PortIdAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, PortIdAttr)
					},
				},
				{
					Name:              EndpointAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return nil
					},
				},
				{
					Name:              ExtractAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						_, interpretationErr := convertExtractorsToDict(true, value)
						return interpretationErr
					},
				},
				{
					Name:              HeadersAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						_, interpretationErr := convertHeadersToMapStringString(true, value)
						return interpretationErr
					},
				},
			},
		},
		Instantiate: instantiateGetHttpRequestRecipe,
	}
}

func instantiateGetHttpRequestRecipe(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(GetHttpRecipeTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &GetHttpRequestRecipe{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type GetHttpRequestRecipe struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
	runtimeValues []string
}

func (recipe *GetHttpRequestRecipe) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := recipe.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &GetHttpRequestRecipe{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (recipe *GetHttpRequestRecipe) Execute(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	service *service.Service,
) (map[string]starlark.Comparable, error) {
	logrus.Debugf("Running get HTTP request recipe '%s'", recipe.String())

	portId, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		recipe.KurtosisValueTypeDefault, PortIdAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Mandatory attribute '%s' was not set on '%s'. This is unexpected and should have been caught earlier", PortIdAttr, GetHttpRecipeTypeName)
	}

	endpoint, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		recipe.KurtosisValueTypeDefault, EndpointAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Mandatory attribute '%s' was not set on '%s'. This is unexpected and should have been caught earlier", EndpointAttr, GetHttpRecipeTypeName)
	}

	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	rawHeaders, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, HeadersAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	headers, interpretationErr := convertHeadersToMapStringString(found, rawHeaders)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	requestResultDict, err := executeInternal(
		ctx,
		serviceNetwork,
		runtimeValueStore,
		service,
		emptyRequestBody,
		portId.GoString(),
		getMethod,
		noContentType,
		endpoint.GoString(),
		extractors,
		headers,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when running HTTP request recipe '%v'", recipe.String())
	}
	return requestResultDict, nil
}

func (recipe *GetHttpRequestRecipe) ResultMapToString(resultMap map[string]starlark.Comparable) string {
	return resultMapToStringInternal(resultMap)
}

func (recipe *GetHttpRequestRecipe) CreateStarlarkReturnValue(resultUuid string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	returnValue, _, interpretationErr := createHttpRequestRecipeStarlarkReturnValueInternal(resultUuid, extractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return returnValue, nil
}

func (recipe *GetHttpRequestRecipe) GetStarlarkReturnValuesAsStringList(resultUuid string) ([]string, *startosis_errors.InterpretationError) {
	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	_, returnValueStrings, interpretationErr := createHttpRequestRecipeStarlarkReturnValueInternal(resultUuid, extractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return returnValueStrings, nil
}

func (recipe *GetHttpRequestRecipe) RequestType() string {
	return getMethod
}

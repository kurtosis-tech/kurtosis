package service_config

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"time"
)

const (
	ReadyConditionsTypeName = "ReadyConditions"

	RecipeAttr    = "recipe"
	FieldAttr     = "field"
	AssertionAttr = "assertion"
	TargetAttr    = "target_value"
	IntervalAttr  = "interval"
	TimeoutAttr   = "timeout"

	defaultInterval = 1 * time.Second
	defaultTimeout  = 15 * time.Minute //TODO we could move these two to the service helpers method
)

func NewReadyConditionsType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ReadyConditionsTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RecipeAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         nil,
				},
				{
					Name:              FieldAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              AssertionAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         assert.ValidateAssertionToken,
				},
				{
					Name:              TargetAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Comparable],
					Validator:         nil,
				},
				{
					Name:              IntervalAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              TimeoutAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
			},
		},
		Instantiate: instantiateReadyConditions,
	}
}

func instantiateReadyConditions(arguments *builtin_argument.ArgumentValuesSet) (kurtosis_type_constructor.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ReadyConditionsTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ReadyConditions{
		KurtosisValueTypeDefault: kurtosisValueType,
		recipe:                   nil,
		field:                    "",
		assertion:                "",
		target:                   starlark.String(""),
		interval:                 defaultInterval,
		timeout:                  defaultTimeout,
	}, nil
}

// ReadyConditions is a starlark.Value that holds all the information needed for ensuring service readiness
type ReadyConditions struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
	recipe    recipe.Recipe
	field     string
	assertion string
	target    starlark.Comparable
	interval  time.Duration
	timeout   time.Duration
}

func (readyConditions *ReadyConditions) ValidateAndFillReadyConditions() *startosis_errors.InterpretationError {

	var (
		genericRecipe     recipe.Recipe
		found             bool
		httpRecipe        *recipe.HttpRequestRecipe
		execRecipe        *recipe.ExecRecipe
		interpretationErr *startosis_errors.InterpretationError
	)

	httpRecipe, found, interpretationErr = kurtosis_type_constructor.ExtractAttrValue[*recipe.HttpRequestRecipe](readyConditions.KurtosisValueTypeDefault, RecipeAttr)
	genericRecipe = httpRecipe
	if interpretationErr != nil || !found {
		execRecipe, found, interpretationErr = kurtosis_type_constructor.ExtractAttrValue[*recipe.ExecRecipe](readyConditions.KurtosisValueTypeDefault, RecipeAttr)
		if interpretationErr != nil {
			return interpretationErr
		}
		if !found {
			return startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
				RecipeAttr, ReadyConditionsTypeName)
		}
		genericRecipe = execRecipe
	}

	field, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyConditions.KurtosisValueTypeDefault, FieldAttr)
	if interpretationErr != nil {
		return interpretationErr
	}
	if !found {
		return startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			FieldAttr, ReadyConditionsTypeName)
	}

	assertion, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyConditions.KurtosisValueTypeDefault, AssertionAttr)
	if interpretationErr != nil {
		return interpretationErr
	}
	if !found {
		return startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			AssertionAttr, ReadyConditionsTypeName)
	}

	target, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Comparable](readyConditions.KurtosisValueTypeDefault, TargetAttr)
	if interpretationErr != nil {
		return interpretationErr
	}
	if !found {
		return startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			TargetAttr, ReadyConditionsTypeName)
	}

	interval := defaultInterval

	intervalStr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyConditions.KurtosisValueTypeDefault, IntervalAttr)
	if interpretationErr != nil {
		return interpretationErr
	}
	if found {
		parsedInterval, parseErr := time.ParseDuration(intervalStr.GoString())
		if parseErr != nil {
			return startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", intervalStr.GoString())
		}
		interval = parsedInterval
	}

	timeout := defaultTimeout

	timeoutStr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyConditions.KurtosisValueTypeDefault, TimeoutAttr)
	if interpretationErr != nil {
		return interpretationErr
	}
	if found {
		parsedTimeout, parseErr := time.ParseDuration(timeoutStr.GoString())
		if parseErr != nil {
			return startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing timeout '%v'", timeoutStr.GoString())
		}
		timeout = parsedTimeout
	}

	readyConditions.recipe = genericRecipe
	readyConditions.field = field.GoString()
	readyConditions.assertion = assertion.GoString()
	readyConditions.target = target
	readyConditions.interval = interval
	readyConditions.timeout = timeout

	return nil
}

func (readyConditions *ReadyConditions) GetRecipe() recipe.Recipe {
	return readyConditions.recipe
}

func (readyConditions *ReadyConditions) GetField() string {
	return readyConditions.field
}

func (readyConditions *ReadyConditions) GetAssertion() string {
	return readyConditions.assertion
}

func (readyConditions *ReadyConditions) GetTarget() starlark.Comparable {
	return readyConditions.target
}

func (readyConditions *ReadyConditions) GetInterval() time.Duration {
	return readyConditions.interval
}

func (readyConditions *ReadyConditions) GetTimeout() time.Duration {
	return readyConditions.timeout
}

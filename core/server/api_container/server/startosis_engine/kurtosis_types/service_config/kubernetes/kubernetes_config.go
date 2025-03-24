package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	KubernetesConfigTypeName       = "KubernetesConfig"
	ExtraIngressConfigAttrTypeName = "extraIngressConfig"
	WorkloadTypeAttrTypeName       = "workload_type"

	// Kubernetes workload types
	WorkloadTypePod        = "pod"
	WorkloadTypeDeployment = "deployment"
)

func NewKubernetesConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: KubernetesConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ExtraIngressConfigAttrTypeName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*ExtraIngressConfig],
				},
				{
					Name:       WorkloadTypeAttrTypeName,
					IsOptional: true,
					ZeroValueProvider: func() starlark.Value {
						// For api backwards compatibility, and forward type consistency,
						// we'll convert unspecified values to an explicit pod request
						return starlark.String(WorkloadTypePod)
					},
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if value == nil {
							// Treat no value as pod for backwards compatibility
							return nil
						}
						workloadType, ok := value.(starlark.String)
						if !ok {
							return startosis_errors.NewInterpretationError("Expected a string value for workload_type")
						}
						if workloadType != WorkloadTypePod && workloadType != WorkloadTypeDeployment {
							return startosis_errors.NewInterpretationError("Invalid workload type '%s'. Allowed values are '%s' or '%s'",
								workloadType, WorkloadTypePod, WorkloadTypeDeployment)
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateKubernetesConfig,
	}
}

func instantiateKubernetesConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(KubernetesConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &KubernetesConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type KubernetesConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *KubernetesConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &KubernetesConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// GetWorkloadType returns the type of Kubernetes workload to use for this service
// Valid values are "", "pod", or "deployment"
// Empty string defaults to "pod"
func (config *KubernetesConfig) GetWorkloadType() (string, *startosis_errors.InterpretationError) {
	workloadTypeStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, WorkloadTypeAttrTypeName)

	if interpretationErr != nil {
		return "", startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred getting the workload_type field")
	}

	// Backwards compatibility support for not specified results in a "raw" pod
	if !found {
		return WorkloadTypePod, nil
	}
	workloadType := string(workloadTypeStarlark)
	if workloadType == "" {
		workloadType = WorkloadTypePod
	}

	// Validate the workload type
	if workloadType != WorkloadTypePod && workloadType != WorkloadTypeDeployment {
		return "", startosis_errors.NewInterpretationError("Invalid workload type '%s'. Allowed values are empty string, '%s', or '%s'",
			workloadType, WorkloadTypePod, WorkloadTypeDeployment)
	}

	return workloadType, nil
}

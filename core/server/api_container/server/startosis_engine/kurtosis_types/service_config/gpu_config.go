package service_config

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	GpuConfigTypeName = "GpuConfig"

	GpuConfigCountAttr     = "count"
	GpuConfigDeviceIDsAttr = "device_ids"
	GpuConfigShmSizeAttr   = "shm_size"
	GpuConfigUlimitsAttr   = "ulimits"
	// GpuConfigDriverAttr accepts either a string shorthand (e.g. "nvidia") or a dict with
	// "docker" and/or "kubernetes" keys for backend-specific overrides.
	GpuConfigDriverAttr = "driver"

	gpuConfigDriverDockerKey     = "docker"
	gpuConfigDriverKubernetesKey = "kubernetes"
)

func NewGpuConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: GpuConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              GpuConfigCountAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator:         nil,
				},
				{
					Name:              GpuConfigDeviceIDsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              GpuConfigShmSizeAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator:         nil,
				},
				{
					Name:              GpuConfigUlimitsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					// ZeroValueProvider returns nil (interface zero value), which causes the
					// framework to skip concrete-type validation and allows string | dict here.
					Name:              GpuConfigDriverAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         nil,
				},
			},
		},
		Instantiate: instantiateGpuConfig,
	}
}

func instantiateGpuConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(GpuConfigTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &GpuConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// GpuConfig is a Starlark value that holds GPU-related configuration for a service.
type GpuConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (g *GpuConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := g.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &GpuConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (g *GpuConfig) ToKurtosisType() (service.GpuConfig, *startosis_errors.InterpretationError) {
	var ok bool

	var count int64
	countStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](g.KurtosisValueTypeDefault, GpuConfigCountAttr)
	if interpretationErr != nil {
		return service.GpuConfig{}, interpretationErr
	}
	if found {
		count, ok = countStarlark.Int64()
		if !ok {
			return service.GpuConfig{}, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to int64", GpuConfigCountAttr, countStarlark)
		}
	}

	var deviceIDs []string
	deviceIDsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](g.KurtosisValueTypeDefault, GpuConfigDeviceIDsAttr)
	if interpretationErr != nil {
		return service.GpuConfig{}, interpretationErr
	}
	if found && deviceIDsStarlark != nil {
		for i := 0; i < deviceIDsStarlark.Len(); i++ {
			strVal, isStr := starlark.AsString(deviceIDsStarlark.Index(i))
			if !isStr {
				return service.GpuConfig{}, startosis_errors.NewInterpretationError("An error occurred parsing field '%v': all elements must be strings", GpuConfigDeviceIDsAttr)
			}
			deviceIDs = append(deviceIDs, strVal)
		}
	}

	var shmSizeMegabytes uint64
	shmSizeStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](g.KurtosisValueTypeDefault, GpuConfigShmSizeAttr)
	if interpretationErr != nil {
		return service.GpuConfig{}, interpretationErr
	}
	if found {
		shmSizeMegabytes, ok = shmSizeStarlark.Uint64()
		if !ok {
			return service.GpuConfig{}, startosis_errors.NewInterpretationError("An error occurred parsing field '%v' with value '%v' to uint64", GpuConfigShmSizeAttr, shmSizeStarlark)
		}
	}

	var ulimits map[string]int64
	ulimitsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](g.KurtosisValueTypeDefault, GpuConfigUlimitsAttr)
	if interpretationErr != nil {
		return service.GpuConfig{}, interpretationErr
	}
	if found && ulimitsStarlark.Len() > 0 {
		ulimits = map[string]int64{}
		for _, item := range ulimitsStarlark.Items() {
			key, keyOk := item[0].(starlark.String)
			if !keyOk {
				return service.GpuConfig{}, startosis_errors.NewInterpretationError("Expected string key in '%v' dict, got '%T'", GpuConfigUlimitsAttr, item[0])
			}
			val, valOk := item[1].(starlark.Int)
			if !valOk {
				return service.GpuConfig{}, startosis_errors.NewInterpretationError("Expected int value in '%v' dict for key '%v', got '%T'", GpuConfigUlimitsAttr, key.GoString(), item[1])
			}
			valInt64, valInt64Ok := val.Int64()
			if !valInt64Ok {
				return service.GpuConfig{}, startosis_errors.NewInterpretationError("Could not convert value for '%v' key '%v' to int64", GpuConfigUlimitsAttr, key.GoString())
			}
			ulimits[key.GoString()] = valInt64
		}
	}

	dockerDriver, k8sResourceName, interpretationErr := parseDriverAttr(g.KurtosisValueTypeDefault)
	if interpretationErr != nil {
		return service.GpuConfig{}, interpretationErr
	}

	return service.NewGpuConfig(count, deviceIDs, shmSizeMegabytes, ulimits, dockerDriver, k8sResourceName), nil
}

// parseDriverAttr extracts the driver field, which can be:
//   - absent → both default to nvidia
//   - a string, e.g. "nvidia" → docker="nvidia", kubernetes="nvidia.com/gpu"
//   - a dict, e.g. {"docker": "amd", "kubernetes": "amd.com/gpu"} → explicit per-backend values
func parseDriverAttr(value *kurtosis_type_constructor.KurtosisValueTypeDefault) (dockerDriver string, k8sResourceName string, interpretationErr *startosis_errors.InterpretationError) {
	driverVal, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](value, GpuConfigDriverAttr)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}
	if !found || driverVal == nil {
		return service.DefaultDockerGpuDriver, service.DefaultK8sGpuResourceName, nil
	}

	switch v := driverVal.(type) {
	case starlark.String:
		s := string(v)
		if s == "" {
			return "", "", startosis_errors.NewInterpretationError("Field '%v' string shorthand cannot be empty", GpuConfigDriverAttr)
		}
		return s, s + ".com/gpu", nil
	case *starlark.Dict:
		return parseDriverDict(v)
	default:
		return "", "", startosis_errors.NewInterpretationError(
			"Field '%v' must be a string or a dict with '%v'/'%v' keys, got '%T'",
			GpuConfigDriverAttr, gpuConfigDriverDockerKey, gpuConfigDriverKubernetesKey, driverVal)
	}
}

func parseDriverDict(d *starlark.Dict) (dockerDriver string, k8sResourceName string, interpretationErr *startosis_errors.InterpretationError) {
	dockerDriver = service.DefaultDockerGpuDriver
	k8sResourceName = service.DefaultK8sGpuResourceName

	if dockerVal, found, err := d.Get(starlark.String(gpuConfigDriverDockerKey)); err == nil && found {
		s, ok := starlark.AsString(dockerVal)
		if !ok || s == "" {
			return "", "", startosis_errors.NewInterpretationError("Key '%v' in field '%v' must be a non-empty string", gpuConfigDriverDockerKey, GpuConfigDriverAttr)
		}
		dockerDriver = s
	}

	if k8sVal, found, err := d.Get(starlark.String(gpuConfigDriverKubernetesKey)); err == nil && found {
		s, ok := starlark.AsString(k8sVal)
		if !ok || s == "" {
			return "", "", startosis_errors.NewInterpretationError("Key '%v' in field '%v' must be a non-empty string", gpuConfigDriverKubernetesKey, GpuConfigDriverAttr)
		}
		k8sResourceName = s
	}

	return dockerDriver, k8sResourceName, nil
}

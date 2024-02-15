package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"reflect"
	"strings"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type APIContainerArgs struct {
	// The version of the API container that was started by the engine, so that the API container can report its
	// own version
	// Ideally this would come from a hardcoded constant, but we don't yet have the machinery that can update a constant
	// on every build
	Version string `json:"version"`

	LogLevel string `json:"logLevel"`

	GrpcListenPortNum uint16 `json:"grpcListenPortNum"`

	EnclaveUUID string `json:"enclaveUuid"`

	// The directory on the API container where the enclave data directory will have been mounted
	EnclaveDataVolumeDirpath string `json:"enclaveDataVolume"`

	KurtosisBackendType KurtosisBackendType `json:"kurtosisBackendType"`

	// Should be deserialized differently depending on value of KurtosisBackendType
	KurtosisBackendConfig interface{} `json:"kurtosisBackendConfig"`

	EnclaveEnvVars string `json:"enclaveEnvVars"`

	IsProductionEnclave bool `json:"isProductionEnclave"`

	//The anonymized user ID for metrics analytics purpose
	MetricsUserID string `json:"metricsUserID"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`

	//If its running in a CI environment
	IsCI bool `json:"is_ci"`

	// The Cloud User ID of the current user if available
	CloudUserID metrics_client.CloudUserID `json:"cloud_user_id"`

	// The Cloud Instance ID of the current user if available
	CloudInstanceID metrics_client.CloudInstanceID `json:"cloud_instance_id"`
}

var skipValidation = map[string]bool{
	"cloud_instance_id": true,
	"cloud_user_id":     true,
}

func (args *APIContainerArgs) UnmarshalJSON(data []byte) error {
	// create a mirror type to avoid unmarshalling infinitely https://stackoverflow.com/questions/52433467/how-to-call-json-unmarshal-inside-unmarshaljson-without-causing-stack-overflow
	type APIContainerArgsMirror APIContainerArgs
	var apiContainerArgsMirror APIContainerArgsMirror
	if err := json.Unmarshal(data, &apiContainerArgsMirror); err != nil {
		return stacktrace.Propagate(err, "Failed to unmarshal api container args")
	}
	byteArray, err := json.Marshal(apiContainerArgsMirror.KurtosisBackendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to re-marshal Kurtosis backend interface ")
	}
	switch apiContainerArgsMirror.KurtosisBackendType {
	case KurtosisBackendType_Docker:
		var dockerConfig kurtosis_backend_config.DockerBackendConfig
		if err := json.Unmarshal(byteArray, &dockerConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", apiContainerArgsMirror.KurtosisBackendConfig, apiContainerArgsMirror.KurtosisBackendType.String())
		}
		apiContainerArgsMirror.KurtosisBackendConfig = dockerConfig
	case KurtosisBackendType_Kubernetes:
		var kubernetesConfig kurtosis_backend_config.KubernetesBackendConfig
		if err := json.Unmarshal(byteArray, &kubernetesConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", apiContainerArgsMirror.KurtosisBackendConfig, apiContainerArgsMirror.KurtosisBackendType.String())
		}
		apiContainerArgsMirror.KurtosisBackendConfig = kubernetesConfig
	default:
		return stacktrace.NewError("Unmarshalled an unrecognized Kurtosis backend type: '%v'", apiContainerArgsMirror.KurtosisBackendType.String())
	}
	*args = APIContainerArgs(apiContainerArgsMirror)
	return nil
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//
//	we get compile errors if there are missing fields
func NewAPIContainerArgs(
	version string,
	logLevel string,
	grpcListenPortNum uint16,
	enclaveUuid string,
	enclaveDataVolumeDirpath string,
	kurtosisBackendType KurtosisBackendType,
	kurtosisBackendConfig interface{},
	enclaveEnvVars string,
	isProductionEnclave bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
) (*APIContainerArgs, error) {
	result := &APIContainerArgs{
		Version:                     version,
		LogLevel:                    logLevel,
		GrpcListenPortNum:           grpcListenPortNum,
		EnclaveUUID:                 enclaveUuid,
		EnclaveDataVolumeDirpath:    enclaveDataVolumeDirpath,
		KurtosisBackendType:         kurtosisBackendType,
		KurtosisBackendConfig:       kurtosisBackendConfig,
		EnclaveEnvVars:              enclaveEnvVars,
		IsProductionEnclave:         isProductionEnclave,
		MetricsUserID:               metricsUserID,
		DidUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		IsCI:                        isCI,
		CloudUserID:                 cloudUserID,
		CloudInstanceID:             cloudInstanceID,
	}

	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating API container args")
	}
	return result, nil
}

// NOTE: We can't use a pointer receiver here else reflection's NumField will panic
func (args APIContainerArgs) validate() error {
	// Generic validation based on field type
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i)
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		if _, found := skipValidation[jsonFieldName]; found {
			continue
		}

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}

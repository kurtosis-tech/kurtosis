package args

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"

	"github.com/kurtosis-tech/kurtosis/engine/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	jsonFieldTag = "json"

	emptyJsonField = "{}"
)

// Fields are public for JSON de/serialization
type EngineServerArgs struct {
	GrpcListenPortNum uint16 `json:"grpcListenPortNum"`

	LogLevelStr string `json:"logLevelStr"`

	// So that the engine server knows its own version
	ImageVersionTag string `json:"imageVersionTag"`

	//The anonymized user ID for metrics analytics purpose `json:"metricsUserId"`
	MetricsUserID string `json:"metricsUserId"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`

	KurtosisBackendType KurtosisBackendType `json:"kurtosisBackendType"`

	// KurtosisLocalBackendConfig corresponds to the config to connect the Kurtosis backend running in the user local
	// laptop. It is mandatory as Kurtosis cannot run without a local backend right now
	// Should be deserialized differently depending on value of KurtosisBackendType
	KurtosisLocalBackendConfig interface{} `json:"kurtosisBackendConfig"`

	// Engine server on bastion host?
	OnBastionHost bool `json:"onBastionHost"`

	// PoolSize represents the enclave pool size, for instance if this value is 3, these amount of idle enclaves
	// will be created when the engine start in order to be used when users request for a new enclave.
	PoolSize uint8 `json:"poolSize"`

	// Environment variable to pass to all the enclaves the engine is going to create. Those environment variable will
	// then be accessible in Starlark scripts in the `kurtosis` module
	EnclaveEnvVars string `json:"enclaveEnvVars"`

	// Whether the Engine is running in a CI  environment
	IsCI bool `json:"is_ci"`

	// The Cloud User ID of the current user if available
	CloudUserID metrics_client.CloudUserID `json:"cloud_user_id"`

	// The Cloud Instance ID of the current user if available
	CloudInstanceID metrics_client.CloudInstanceID `json:"cloud_instance_id"`

	// List of allowed origins to validate CORS requests on the REST API. If undefined, defaults to '*' (any origin).
	AllowedCORSOrigins *[]string `json:"allowed_cors_origins,omitempty"`

	// To restart the current API containers after the engine has been restarted
	RestartAPIContainers bool `json:"restart_api_containers"`

	// Enclave manager UI domain name
	Domain string `json:"domain"`

	LogRetentionPeriod string `json:"logRetentionPeriod"`
}

var skipValidation = map[string]bool{
	"cloud_instance_id": true,
	"cloud_user_id":     true,
	"domain":            true,
}

func (args *EngineServerArgs) UnmarshalJSON(data []byte) error {
	// create a mirror type to avoid unmarshalling infinitely https://stackoverflow.com/questions/52433467/how-to-call-json-unmarshal-inside-unmarshaljson-without-causing-stack-overflow
	type EngineServerArgsMirror EngineServerArgs
	var engineServerArgsMirror EngineServerArgsMirror
	if err := json.Unmarshal(data, &engineServerArgsMirror); err != nil {
		return stacktrace.Propagate(err, "Failed to unmarshal engine server args")
	}
	byteArray, err := json.Marshal(engineServerArgsMirror.KurtosisLocalBackendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to re-marshal interface ")
	}
	switch engineServerArgsMirror.KurtosisBackendType {
	case KurtosisBackendType_Docker:
		var dockerConfig kurtosis_backend_config.DockerBackendConfig
		if err := json.Unmarshal(byteArray, &dockerConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", engineServerArgsMirror.KurtosisLocalBackendConfig, engineServerArgsMirror.KurtosisBackendType.String())
		}
		engineServerArgsMirror.KurtosisLocalBackendConfig = dockerConfig
	case KurtosisBackendType_Kubernetes:
		var kubernetesConfig kurtosis_backend_config.KubernetesBackendConfig
		if err := json.Unmarshal(byteArray, &kubernetesConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", engineServerArgsMirror.KurtosisLocalBackendConfig, engineServerArgsMirror.KurtosisBackendType.String())
		}
		engineServerArgsMirror.KurtosisLocalBackendConfig = kubernetesConfig
	default:
		return stacktrace.NewError("Unmarshalled an unrecognized Kurtosis backend type: '%v'", engineServerArgsMirror.KurtosisBackendType.String())
	}
	*args = EngineServerArgs(engineServerArgsMirror)
	return nil
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
// we get compile errors if there are missing fields
func NewEngineServerArgs(
	grpcListenPortNum uint16,
	logLevelStr string,
	imageVersionTag string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	kurtosisBackendType KurtosisBackendType,
	kurtosisLocalBackendConfig interface{},
	onBastionHost bool,
	poolSize uint8,
	enclaveEnvVars string,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
	allowedCORSOrigins *[]string,
	restartAPIContainers bool,
	domain string,
	logRetentionPeriod string,
) (*EngineServerArgs, error) {
	if enclaveEnvVars == "" {
		enclaveEnvVars = emptyJsonField
	}
	result := &EngineServerArgs{
		GrpcListenPortNum:           grpcListenPortNum,
		LogLevelStr:                 logLevelStr,
		ImageVersionTag:             imageVersionTag,
		MetricsUserID:               metricsUserID,
		DidUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		KurtosisBackendType:         kurtosisBackendType,
		KurtosisLocalBackendConfig:  kurtosisLocalBackendConfig,
		OnBastionHost:               onBastionHost,
		PoolSize:                    poolSize,
		EnclaveEnvVars:              enclaveEnvVars,
		IsCI:                        isCI,
		CloudUserID:                 cloudUserID,
		CloudInstanceID:             cloudInstanceID,
		AllowedCORSOrigins:          allowedCORSOrigins,
		RestartAPIContainers:        restartAPIContainers,
		Domain:                      domain,
		LogRetentionPeriod:          logRetentionPeriod,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating engine server args")
	}
	return result, nil
}

func (args EngineServerArgs) validate() error {
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

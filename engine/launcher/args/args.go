package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/remote_context_backend"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args/kurtosis_backend_config"
	"reflect"
	"strings"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type EngineServerArgs struct {
	GrpcListenPortNum      uint16 `json:"grpcListenPortNum"`
	GrpcProxyListenPortNum uint16 `json:"grpcProxyListenPortNum"`

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

	// KurtosisRemoteBackendConfig corresponds to the config to connect to an optional remote backend. It is used only
	// when Kurtosis is using a dual-backend context (both local and remote). In this case, Kurtosis connects  to both
	// the local backend and the remote backend using this configuration and the above KurtosisLocalBackendConfig
	// Is nil when Kurtosis is used in a local-only context
	KurtosisRemoteBackendConfig *remote_context_backend.KurtosisRemoteBackendConfig `json:"kurtosisRemoteBackendConfig,omitempty"`
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
	grpcProxyListenPortNum uint16,
	logLevelStr string,
	imageVersionTag string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	kurtosisBackendType KurtosisBackendType,
	kurtosisLocalBackendConfig interface{},
	kurtosisRemoteBackendConfig *remote_context_backend.KurtosisRemoteBackendConfig,
) (*EngineServerArgs, error) {
	result := &EngineServerArgs{
		GrpcListenPortNum:           grpcListenPortNum,
		GrpcProxyListenPortNum:      grpcProxyListenPortNum,
		LogLevelStr:                 logLevelStr,
		ImageVersionTag:             imageVersionTag,
		MetricsUserID:               metricsUserID,
		DidUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		KurtosisBackendType:         kurtosisBackendType,
		KurtosisLocalBackendConfig:  kurtosisLocalBackendConfig,
		KurtosisRemoteBackendConfig: kurtosisRemoteBackendConfig,
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

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}

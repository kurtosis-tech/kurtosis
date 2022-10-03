package args

import (
	"encoding/json"
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
	GrpcListenPortNum        uint16 `json:"grpcListenPortNum"`
	GrpcProxyListenPortNum   uint16 `json:"grpcProxyListenPortNum"`
	LogsCollectorHttpPortNum uint16 `json:"logsCollectorHttpPortNum"`

	LogLevelStr string `json:"logLevelStr"`

	// So that the engine server knows its own version
	ImageVersionTag string `json:"imageVersionTag"`

	//The anonymized user ID for metrics analytics purpose `json:"metricsUserId"`
	MetricsUserID string `json:"metricsUserId"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`

	KurtosisBackendType KurtosisBackendType `json:"kurtosisBackendType"`

	// Should be deserialized differently depending on value of KurtosisBackendType
	KurtosisBackendConfig interface{} `json:"kurtosisBackendConfig"`
}

func (args *EngineServerArgs) UnmarshalJSON(data []byte) error {
	// create a mirror type to avoid unmarshalling infinitely https://stackoverflow.com/questions/52433467/how-to-call-json-unmarshal-inside-unmarshaljson-without-causing-stack-overflow
	type EngineServerArgsMirror EngineServerArgs
	var engineServerArgsMirror EngineServerArgsMirror
	if err := json.Unmarshal(data, &engineServerArgsMirror); err != nil {
		return stacktrace.Propagate(err, "Failed to unmarshal engine server args")
	}
	byteArray, err := json.Marshal(engineServerArgsMirror.KurtosisBackendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to re-marshal interface ")
	}
	switch engineServerArgsMirror.KurtosisBackendType {
	case KurtosisBackendType_Docker:
		var dockerConfig kurtosis_backend_config.DockerBackendConfig
		if err := json.Unmarshal(byteArray, &dockerConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", engineServerArgsMirror.KurtosisBackendConfig, engineServerArgsMirror.KurtosisBackendType.String())
		}
		engineServerArgsMirror.KurtosisBackendConfig = dockerConfig
	case KurtosisBackendType_Kubernetes:
		var kubernetesConfig kurtosis_backend_config.KubernetesBackendConfig
		if err := json.Unmarshal(byteArray, &kubernetesConfig); err != nil {
			return stacktrace.Propagate(err, "Failed to unmarshal backend config '%+v' with type '%v'", engineServerArgsMirror.KurtosisBackendConfig, engineServerArgsMirror.KurtosisBackendType.String())
		}
		engineServerArgsMirror.KurtosisBackendConfig = kubernetesConfig
	default:
		return stacktrace.NewError("Unmarshalled an unrecognized Kurtosis backend type: '%v'", engineServerArgsMirror.KurtosisBackendType.String())
	}
	*args = EngineServerArgs(engineServerArgsMirror)
	return nil
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewEngineServerArgs(
	grpcListenPortNum uint16,
	grpcProxyListenPortNum uint16,
	logLevelStr string,
	imageVersionTag string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	kurtosisBackendType KurtosisBackendType,
	kurtosisBackendConfig interface{},
) (*EngineServerArgs, error) {
	result := &EngineServerArgs{
		GrpcListenPortNum:           grpcListenPortNum,
		GrpcProxyListenPortNum:      grpcProxyListenPortNum,
		LogLevelStr:                 logLevelStr,
		ImageVersionTag:             imageVersionTag,
		MetricsUserID:               metricsUserID,
		DidUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		KurtosisBackendType:         kurtosisBackendType,
		KurtosisBackendConfig:       kurtosisBackendConfig,
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

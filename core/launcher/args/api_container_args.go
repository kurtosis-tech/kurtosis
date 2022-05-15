package args

import (
	"github.com/kurtosis-tech/kurtosis-core/launcher/args/kurtosis_backend_type"
	"reflect"
	"strings"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type APIContainerArgs struct {
	Version string `json:"version"`

	LogLevel string `json:"logLevel"`

	GrpcListenPortNum      uint16 `json:"grpcListenPortNum"`
	GrpcProxyListenPortNum uint16 `json:"grpcProxyListenPortNum"`

	EnclaveId  string `json:"enclaveId"`

	IsPartitioningEnabled bool `json:"isPartitioningEnabled"`

	//The anonymized user ID for metrics analytics purpose
	MetricsUserID string `json:"metricsUserID"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`

	// The directory on the API container where the enclave data directory will have been mounted
	EnclaveDataVolumeDirpath string `json:"enclaveDataVolume"`

	KurtosisBackendType kurtosis_backend_type.KurtosisBackendType `json:"kurtosisBackendType"`

	// Should be deserialized differently depending on value of KurtosisBackendType
	KurtosisBackendConfig interface{} `json:"kurtosisBackendConfig"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewAPIContainerArgs(
	version string,
	logLevel string,
	grpcListenPortNum uint16,
	grpcProxyListenPortNum uint16,
	enclaveId string,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	enclaveDataVolumeDirpath string,
	kurtosisBackendType kurtosis_backend_type.KurtosisBackendType,
	kurtosisBackendConfig interface{},
) (*APIContainerArgs, error) {
	result := &APIContainerArgs{
		Version:                     version,
		LogLevel:                    logLevel,
		GrpcListenPortNum:           grpcListenPortNum,
		GrpcProxyListenPortNum:      grpcProxyListenPortNum,
		EnclaveId:                   enclaveId,
		IsPartitioningEnabled:       isPartitioningEnabled,
		MetricsUserID:               metricsUserID,
		DidUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		EnclaveDataVolumeDirpath:    enclaveDataVolumeDirpath,
		KurtosisBackendType:         kurtosisBackendType,
		KurtosisBackendConfig:       kurtosisBackendConfig,
	}

	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating API container args")
	}
	return result, nil
}

func (args APIContainerArgs) validate() error {
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

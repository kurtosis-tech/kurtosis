package args

import (
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
	NetworkId  string `json:"networkId"`
	SubnetMask string `json:"subnetMask"`

	// Necessary so that when the API container starts modules, it knows which IP addr to give them
	ApiContainerIpAddr string `json:"apiContainerIpAddr"`

	// Instructs the API container that these IP addrs are already taken and shouldn't be used
	TakenIpAddrs map[string]bool `json:"takenIpAddrsSet"`

	IsPartitioningEnabled bool `json:"isPartitioningEnabled"`

	// TODO Remove when we've verified enclave data volume is working
	// The location on the API container where the enclave data directory will have been bind-mounted
	EnclaveDataDirpathOnAPIContainer string `json:"enclaveDataDirpathOnAPIContainer"`

	// TODO Remove when we've verified enclave data volume is working
	// The dirpath on the Docker host machine where enclave data is stored, which the API container
	//  will use to bind-mount the directory into the services that it starts
	EnclaveDataDirpathOnHostMachine string `json:"enclaveDataDirpathOnHostMachine"`

	//The anonymized user ID for metrics analytics purpose
	MetricsUserID string `json:"metricsUserID"`

	//User consent to send metrics
	DidUserAcceptSendingMetrics bool `json:"didUserAcceptSendingMetrics"`

	// The directory on the API container where the enclave data directory will have been mounted
	EnclaveDataVolumeDirpath string `json:"enclaveDataVolume"`

	// KurtosisBackend configuration
	KurtosisBackendType string `json:"backendType"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewAPIContainerArgs(
	version string,
	logLevel string,
	grpcListenPortNum uint16,
	grpcProxyListenPortNum uint16,
	enclaveId string,
	networkId string,
	subnetMask string,
	apiContainerIpAddr string,
	takenIpAddrs map[string]bool,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnAPIContainer string,
	enclaveDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	enclaveDataVolumeDirpath string,
	kurtosisBackendType string,
) (*APIContainerArgs, error) {
	result := &APIContainerArgs{
		Version:                          version,
		LogLevel:                         logLevel,
		GrpcListenPortNum:                grpcListenPortNum,
		GrpcProxyListenPortNum:           grpcProxyListenPortNum,
		EnclaveId:                        enclaveId,
		NetworkId:                        networkId,
		SubnetMask:                       subnetMask,
		ApiContainerIpAddr:               apiContainerIpAddr,
		TakenIpAddrs:                     takenIpAddrs,
		IsPartitioningEnabled:            isPartitioningEnabled,
		EnclaveDataDirpathOnAPIContainer: enclaveDataDirpathOnAPIContainer,
		EnclaveDataDirpathOnHostMachine:  enclaveDataDirpathOnHostMachine,
		MetricsUserID:                    metricsUserID,
		DidUserAcceptSendingMetrics:      didUserAcceptSendingMetrics,
		EnclaveDataVolumeDirpath:         enclaveDataVolumeDirpath,
		KurtosisBackendType: 			  kurtosisBackendType,
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

package args

import (
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
	"strings"
)

const (
	jsonFieldTag          = "json"
)

// Fields are public for JSON de/serialization
type APIContainerArgs struct {
	// The name of the API container itself (will be used to get its own container ID)
	ContainerName			 string	`json:"containerName"`

	LogLevel                 string `json:"logLevel"`

	ListenPortNum      uint16 `json:"listenPortNum"`
	ListenPortProtocol string `json:"listenPortProtocol"`

	EnclaveId				 string `json:"enclaveId"'`
	NetworkId                string `json:"networkId"`
	SubnetMask               string	`json:"subnetMask"`

	// Necessary so that when the API container starts modules, it knows which IP addr to give them
	ApiContainerIpAddr string	`json:"apiContainerIpAddr"`

	// Instructs the API container that these IP addrs are already taken and shouldn't be used
	TakenIpAddrs			 map[string]bool `json:"takenIpAddrsSet"`

	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`

	// Whether the ports of the containers that the API service starts should be published to the Docker host machine
	ShouldPublishPorts bool		`json:"shouldPublishPorts"`

	// The location on the API container where the enclave data directory will have been bind-mounted
	EnclaveDataDirpathOnAPIContainer string `json:"enclaveDataDirpathOnAPIContainer"`

	// The dirpath on the Docker host machine where enclave data is stored, which the API container
	//  will use to bind-mount the directory into the services that it starts
	EnclaveDataDirpathOnHostMachine string	`json:"enclaveDataDirpathOnHostMachine"`
}


// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewAPIContainerArgs(
	containerName string,
	logLevel string,
	listenPortNum uint16,
	listenPortProtocol string,
	enclaveId string,
	networkId string,
	subnetMask string,
	apiContainerIpAddr string,
	takenIpAddrs map[string]bool,
	isPartitioningEnabled bool,
	shouldPublishPorts bool,
	enclaveDataDirpathOnAPIContainer string,
	enclaveDataDirpathOnHostMachine string,
) (*APIContainerArgs, error) {
	result := &APIContainerArgs{
		ContainerName:                    containerName,
		LogLevel:                         logLevel,
		ListenPortNum:                    listenPortNum,
		ListenPortProtocol:               listenPortProtocol,
		EnclaveId:                        enclaveId,
		NetworkId:                        networkId,
		SubnetMask:                       subnetMask,
		ApiContainerIpAddr:               apiContainerIpAddr,
		TakenIpAddrs:                     takenIpAddrs,
		IsPartitioningEnabled:            isPartitioningEnabled,
		ShouldPublishPorts:               shouldPublishPorts,
		EnclaveDataDirpathOnAPIContainer: enclaveDataDirpathOnAPIContainer,
		EnclaveDataDirpathOnHostMachine:  enclaveDataDirpathOnHostMachine,
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
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}


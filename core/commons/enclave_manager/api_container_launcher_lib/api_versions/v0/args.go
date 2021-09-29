package v0

import (
	"github.com/palantir/stacktrace"
	"reflect"
	"strings"
)

const (
	jsonFieldTag          = "json"
)

// Fields are public for JSON de/serialization
type V0LaunchAPIArgs struct {
	// The name of the API container itself (will be used to get its own container ID)
	ContainerName			 string	`json:"containerName"`

	ContainerLabels          map[string] string `json:"containerLabels"`

	LogLevel                 string `json:"logLevel"`

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

	// When shutting down, the API container will tear down everything in the enclave network that it knows about
	// However, there can be containers which have been connected to the network which should NOT be stopped (e.g. the
	// testing framework initializer container), which should merely be disconnected from the network
	// This is a set of those containers
	ExternalMountedContainerIds	map[string]bool 	`json:"externalMountedContainerIds"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func newV0LaunchAPIArgs(
		containerName string,
		containerLabels map[string]string,
		logLevel string,
		enclaveId string,
		networkId string,
		subnetMask string,
		apiContainerIpAddr string,
		takenIpAddrs map[string]bool,
		isPartitioningEnabled bool,
		shouldPublishPorts bool,
		externalMountedContainerIds map[string]bool) *V0LaunchAPIArgs {
	return &V0LaunchAPIArgs{
		ContainerName:               containerName,
		ContainerLabels:             containerLabels,
		LogLevel:                    logLevel,
		EnclaveId:                   enclaveId,
		NetworkId:                   networkId,
		SubnetMask:                  subnetMask,
		ApiContainerIpAddr:          apiContainerIpAddr,
		TakenIpAddrs:                takenIpAddrs,
		IsPartitioningEnabled:       isPartitioningEnabled,
		ShouldPublishPorts:          shouldPublishPorts,
		ExternalMountedContainerIds: externalMountedContainerIds,
	}
}

func (args V0LaunchAPIArgs) Validate() error {
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


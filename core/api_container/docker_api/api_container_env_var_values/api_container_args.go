/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_env_var_values

import (
	"github.com/palantir/stacktrace"
	"reflect"
	"strings"
)

const (
	jsonFieldTag          = "json"
)

// Params that the API container will use to construct the free host port binding supplier that it will use
// for binding service ports to host ports for debugging purposes
type HostPortBindingSupplierParams struct {
	// The host interface IP that the doled-out ports should be on
	InterfaceIp		string `json:"interfaceIp"`

	// The protocol of the ports that the supplier will produce
	Protocol		string `json:"protocol"`

	// Inclusive start of port range that will be doled out
	PortRangeStart	uint32 `json:"portRangeStart"`

	// EXCLUSIVE end of port range that will be doled out
	PortRangeEnd 	uint32	`json:"portRangeEnd"`

	// "Set" of ports that are already taken and shouldn't be doled out by the free host port binding supplier
	// NOTE: This is more a hint, rather than a requirement, since the FreeHostPortBindingSupplier that's constructed
	//  from these params will test if a port is free regardless
	TakenPorts map[uint32]bool	`json:"takenPorts"`
}

// Even though the fields are public, we create & use this constructor so that new fields will cause an API break
func NewHostPortBindingSupplierParams(interfaceIp string, protocol string, portRangeStart uint32, portRangeEnd uint32, takenPorts map[uint32]bool) *HostPortBindingSupplierParams {
	return &HostPortBindingSupplierParams{InterfaceIp: interfaceIp, Protocol: protocol, PortRangeStart: portRangeStart, PortRangeEnd: portRangeEnd, TakenPorts: takenPorts}
}

// Fields are public for JSON de/serialization
type ApiContainerArgs struct {
	ExecutionInstanceId      string	`json:"executionInstanceId"`
	NetworkId                string `json:"networkId"`
	SubnetMask               string	`json:"subnetMask"`
	GatewayIpAddr            string	`json:"gatewayIpAddr"`
	EnclaveDataVolumeName 	 string	`json:"enclaveDataVolumeName"`

	// Name elements used for identifying the enclave in which this instance of the API container is
	//  running. For example, when running a test, this will be [execution_instance_uuid, test_name]
	EnclaveNameElems []string `json:"enclaveNameElems"`

	// Necessary so that when the API container starts modules, it knows which IP addr to give them
	ApiContainerIpAddr string

	// Instructs the API container that these IP addrs are already taken and shouldn't be used
	TakenIpAddrs			 map[string]bool `json:"takenIpAddrsSet"`

	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`

	// A non-nil value indicates that the Kurtosis API container should bind service ports to ports on the
	// host machine running Kurtosis, for debugging purposes
	// If this is nil, no host port binding will occur
	HostPortBindingSupplierParams *HostPortBindingSupplierParams `json:"hostPortBindingSupplierParams"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewApiContainerArgs(
		executionInstanceId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr string,
		enclaveDataVolumeName string,
		enclaveNameElems []string,
		apiContainerIpAddr string,
		takenIpAddrs map[string]bool,
		isPartitioningEnabled bool,
		hostPortBindingSupplierParams *HostPortBindingSupplierParams) (*ApiContainerArgs, error) {
	result := ApiContainerArgs{
		ExecutionInstanceId:           executionInstanceId,
		NetworkId:                     networkId,
		SubnetMask:                    subnetMask,
		GatewayIpAddr:                 gatewayIpAddr,
		EnclaveDataVolumeName:      enclaveDataVolumeName,
		EnclaveNameElems:              enclaveNameElems,
		ApiContainerIpAddr:            apiContainerIpAddr,
		TakenIpAddrs:                  takenIpAddrs,
		IsPartitioningEnabled:         isPartitioningEnabled,
		HostPortBindingSupplierParams: hostPortBindingSupplierParams,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating test execution args")
	}
	return &result, nil
}

func (args ApiContainerArgs) validate() error {
	if len(args.EnclaveNameElems) == 0 {
		return stacktrace.NewError("Length of enclave name elems must be > 0")
	}

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


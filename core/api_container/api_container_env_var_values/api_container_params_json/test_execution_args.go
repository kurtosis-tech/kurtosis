/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_params_json

import (
	"github.com/palantir/stacktrace"
	"reflect"
	"strings"
)

const (
	jsonFieldTag          = "json"
	setupTimeoutFieldname = "testSetupTimeout"
	runTimeoutFieldname   = "testRunTimeout"
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
	TakenPorts map[uint32]bool	`json:"takenPorts"`
}

// Fields are public for JSON de/serialization
type TestExecutionArgs struct {
	ExecutionInstanceId      string	`json:"executionInstanceId"`
	NetworkId                string `json:"networkId"`
	SubnetMask               string	`json:"subnetMask"`
	GatewayIpAddr            string	`json:"gatewayIpAddr"`
	TestName                 string	`json:"testName"`
	SuiteExecutionVolumeName string	`json:"suiteExecutionVolumeName"`
	TestSuiteContainerId     string	`json:"testSuiteContainerId"`

	// It seems weird that we require this given that the test suite container doesn't run a server, but it's only so
	//  that our free IP address tracker knows not to dole out the test suite container's IP address
	TestSuiteContainerIpAddr string	`json:"testSuiteContainerIpAddr"`
	ApiContainerIpAddr       string	`json:"apiContainerIpAddr"`

	TestSetupTimeoutInSeconds uint32 `json:"testSetupTimeout"`
	TestRunTimeoutInSeconds   uint32 `json:"testRunTimeout"`

	// TODO remove this by passing over the test metadata as one of the params, so that the API container
	//  knows about the metadata
	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`

	// A non-nil value indicates that the Kurtosis API container should bind service ports to ports on the
	// host machine running Kurtosis, for debugging purposes
	// If this is nil, no host port binding will occur
	HostPortBindingSupplierParams *HostPortBindingSupplierParams `json:"hostPortBindingSupplierParams"`
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewTestExecutionArgs(
		executionInstanceId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr string,
		testName string,
		suiteExecutionVolumeName string,
		testSuiteContainerId string,
		testSuiteContainerIpAddr string,
		apiContainerIpAddr string,
		testSetupTimeoutInSeconds uint32,
		testRunTimeoutInSeconds uint32,
		isPartitioningEnabled bool,
		hostPortBindingSupplierParams *HostPortBindingSupplierParams) (*TestExecutionArgs, error) {
	result := TestExecutionArgs{
		ExecutionInstanceId:       executionInstanceId,
		NetworkId:                 networkId,
		SubnetMask:                subnetMask,
		GatewayIpAddr:             gatewayIpAddr,
		TestName:                  testName,
		SuiteExecutionVolumeName:  suiteExecutionVolumeName,
		TestSuiteContainerId:      testSuiteContainerId,
		TestSuiteContainerIpAddr:  testSuiteContainerIpAddr,
		ApiContainerIpAddr:        apiContainerIpAddr,
		TestSetupTimeoutInSeconds: testSetupTimeoutInSeconds,
		TestRunTimeoutInSeconds:   testRunTimeoutInSeconds,
		IsPartitioningEnabled:     isPartitioningEnabled,
		HostPortBindingSupplierParams: hostPortBindingSupplierParams,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating test execution args")
	}
	return &result, nil
}

func (args TestExecutionArgs) validate() error {
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)
		if jsonFieldName == setupTimeoutFieldname || jsonFieldName == runTimeoutFieldname {
			intVal := reflectVal.Field(i).Uint()
			if intVal <= 0 {
				return stacktrace.NewError("JSON field '%s' representing a timeout has value %v, but it must be greater than 0", jsonFieldName, intVal)
			}
		}
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}


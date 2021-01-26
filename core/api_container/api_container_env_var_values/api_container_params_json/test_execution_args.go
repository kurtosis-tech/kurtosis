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
	jsonFieldTag = "json"
)

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

	// TODO remove this by passing over the test metadata as one of the params, so that the API container
	//  knows about the metadata
	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`
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
		isPartitioningEnabled bool) (*TestExecutionArgs, error) {
	result := TestExecutionArgs{
		ExecutionInstanceId: executionInstanceId,
		NetworkId: networkId,
		SubnetMask: subnetMask,
		GatewayIpAddr: gatewayIpAddr,
		TestName: testName,
		SuiteExecutionVolumeName: suiteExecutionVolumeName,
		TestSuiteContainerId: testSuiteContainerId,
		TestSuiteContainerIpAddr: testSuiteContainerIpAddr,
		ApiContainerIpAddr: apiContainerIpAddr,
		IsPartitioningEnabled: isPartitioningEnabled,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating test execution args")
	}
	return &result, nil
}

func (args TestExecutionArgs) validate() error {
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	reflectValElem := reflectVal.Elem()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)
		strVal := reflectValElem.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}


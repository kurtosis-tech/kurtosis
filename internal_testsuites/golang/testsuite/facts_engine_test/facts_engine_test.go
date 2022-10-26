package facts_engine_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testName              = "facts-engine"
	isPartitioningEnabled = false

	factsEngineTestImage = "httpd:2.4.54"
	containerPortId      = "port_id"

	inputForExecFactTest                   = "hello"
	expectedOutputForConstantFactOutput    = "value"
	expectedOutputForExecFactOutput        = "hello\n"
	expectedOutputForHttpRequestFactOutput = "<html><body><h1>It works!</h1></body></html>\n"
	testServiceId                          = "test"
	constantFactName                       = "constant_fact"
	httpRequestFactName                    = "http_request_fact"
	execFactName                           = "exec_fact"
)

var containerUsedPorts = map[string]*services.PortSpec{
	containerPortId: services.NewPortSpec(80, services.PortProtocol_TCP),
}

var execCommandThatShouldWork = []string{
	"true",
}

var execCommandThatShouldHaveLogOutput = []string{
	"echo",
	inputForExecFactTest,
}

func TestExecCommand(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfig := getContainerConfig()

	_, err = enclaveCtx.AddService(testServiceId, containerConfig)
	require.NoError(t, err, "An error occurred starting service '%v'", testServiceId)
	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Defining constant fact...")
	_, err = enclaveCtx.DefineFact(&kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: testServiceId,
		FactName:  constantFactName,
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: &kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
				FactValue: &kurtosis_core_rpc_api_bindings.FactValue{
					FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
						StringValue: expectedOutputForConstantFactOutput,
					},
				},
			},
		},
	})
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	logrus.Infof("Getting constant fact value...")
	getFactValuesResponse, err := enclaveCtx.GetFactValues(testServiceId, constantFactName)
	require.Equal(t, expectedOutputForConstantFactOutput, getFactValuesResponse.GetFactValues()[0].GetStringValue())

	logrus.Infof("Defining exec fact...")
	_, err = enclaveCtx.DefineFact(&kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: testServiceId,
		FactName:  execFactName,
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_ExecFact{
			ExecFact: &kurtosis_core_rpc_api_bindings.ExecFactRecipe{
				CmdArgs: execCommandThatShouldHaveLogOutput,
			},
		},
	})
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	logrus.Infof("Getting exec fact value...")
	getFactValuesResponse, err = enclaveCtx.GetFactValues(testServiceId, execFactName)
	require.Equal(t, expectedOutputForExecFactOutput, getFactValuesResponse.GetFactValues()[0].GetStringValue())

	logrus.Infof("Defining HTTP request fact...")
	_, err = enclaveCtx.DefineFact(&kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: testServiceId,
		FactName:  httpRequestFactName,
		FactRecipe: &kurtosis_core_rpc_api_bindings.FactRecipe_HttpRequestFact{
			HttpRequestFact: &kurtosis_core_rpc_api_bindings.HttpRequestFactRecipe{
				PortId:   containerPortId,
				Method:   kurtosis_core_rpc_api_bindings.HttpRequestMethod_GET,
				Endpoint: "/",
			},
		},
	})
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	logrus.Infof("Getting HTTP request fact value...")
	getFactValuesResponse, err = enclaveCtx.GetFactValues(testServiceId, httpRequestFactName)
	require.Nil(t, err)
	require.Equal(t, expectedOutputForHttpRequestFactOutput, getFactValuesResponse.GetFactValues()[0].GetStringValue())

}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func getContainerConfig() *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(factsEngineTestImage).WithUsedPorts(containerUsedPorts).Build()
	return containerConfig
}

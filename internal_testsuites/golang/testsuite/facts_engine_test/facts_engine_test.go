package facts_engine_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
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
	containerPortId      = "port-id"

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
	constantFactRecipe := binding_constructors.NewConstantFactRecipeWithDefaultRefresh(
		testServiceId,
		constantFactName,
		&kurtosis_core_rpc_api_bindings.ConstantFactRecipe{
			FactValue: &kurtosis_core_rpc_api_bindings.FactValue{
				FactValue: &kurtosis_core_rpc_api_bindings.FactValue_StringValue{
					StringValue: expectedOutputForConstantFactOutput,
				},
			},
		})
	_, err = enclaveCtx.DefineFact(constantFactRecipe)
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	logrus.Infof("Getting constant fact value...")
	getFactValuesResponse, err := enclaveCtx.GetFactValues(testServiceId, constantFactName)
	require.Nil(t, err)
	require.Equal(t, expectedOutputForConstantFactOutput, getFactValuesResponse.GetFactValues()[0].GetStringValue())

	logrus.Infof("Defining exec fact...")
	_, err = enclaveCtx.DefineFact(binding_constructors.NewExecFactRecipeWithDefaultRefresh(testServiceId, execFactName, execCommandThatShouldHaveLogOutput))
	require.Nil(t, err)
	time.Sleep(5 * time.Second)

	logrus.Infof("Getting exec fact value...")
	getFactValuesResponse, err = enclaveCtx.GetFactValues(testServiceId, execFactName)
	require.Nil(t, err)
	require.Equal(t, expectedOutputForExecFactOutput, getFactValuesResponse.GetFactValues()[0].GetStringValue())

	logrus.Infof("Defining HTTP request fact...")
	_, err = enclaveCtx.DefineFact(binding_constructors.NewGetHttpRequestFactRecipeWithDefaultRefresh(testServiceId, httpRequestFactName, containerPortId, "/"))
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

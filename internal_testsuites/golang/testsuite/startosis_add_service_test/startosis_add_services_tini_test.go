package startosis_add_service_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	addServicesWithTiniDisabled = `
def run(plan):
    plan.add_service(name = 'wireguard', config = ServiceConfig(image="linuxserver/wireguard",ports={"docker-51820": PortSpec(number=51820,transport_protocol="UDP",wait="30s")},env_vars={"PUID": "1000","PGID": "1000","TZ": "Etc/UTC"}, tini_enabled=False))
`
)

func (suite *StartosisAddServiceTestSuite) TestAddServicesTiniDisabled() {
	ctx := context.Background()
	runResult, err := suite.RunScript(ctx, addServicesWithTiniDisabled)

	t := suite.T()

	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Service 'wireguard' added with service UUID '[a-f0-9]{32}'`
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script")

	// Ensure that the service is listed
	expectedNumberOfServices := 1
	serviceInfos, err := suite.enclaveCtx.GetServices()
	require.Nil(t, err)
	actualNumberOfServices := len(serviceInfos)
	require.Equal(t, expectedNumberOfServices, actualNumberOfServices)
}

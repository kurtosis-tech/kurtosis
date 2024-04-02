package add

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsePortSpecstr_SuccessCases(t *testing.T) {
	type args struct {
		specStr string
	}

	parsePortSpecSuccessTests := []struct {
		name string
		args args
		want *kurtosis_core_rpc_api_bindings.Port
	}{
		{
			name: "Successfully parse str with application protocol and without transport protocol",
			args: args{
				specStr: "http:3333",
			},
			want: &kurtosis_core_rpc_api_bindings.Port{Number: uint32(3333), TransportProtocol: kurtosis_core_rpc_api_bindings.Port_TCP, MaybeApplicationProtocol: "http", MaybeWaitTimeout: defaultPortWaitTimeoutStr, Locked: nil},
		},
		{
			name: "Successfully parse str with application protocol and with transport protocol",
			args: args{
				specStr: "http:3333/udp",
			},
			want: &kurtosis_core_rpc_api_bindings.Port{Number: uint32(3333), TransportProtocol: kurtosis_core_rpc_api_bindings.Port_UDP, MaybeApplicationProtocol: "http", MaybeWaitTimeout: defaultPortWaitTimeoutStr, Locked: nil},
		},
		{
			name: "Successfully parse str without application protocol and with transport protocol",
			args: args{
				specStr: "3333/udp",
			},
			want: &kurtosis_core_rpc_api_bindings.Port{Number: uint32(3333), TransportProtocol: kurtosis_core_rpc_api_bindings.Port_UDP, MaybeApplicationProtocol: "", MaybeWaitTimeout: defaultPortWaitTimeoutStr, Locked: nil},
		},
		{
			name: "Successfully parse str without application protocol and without transport protocol",
			args: args{
				specStr: "3333",
			},
			want: &kurtosis_core_rpc_api_bindings.Port{Number: uint32(3333), TransportProtocol: kurtosis_core_rpc_api_bindings.Port_TCP, MaybeApplicationProtocol: "", MaybeWaitTimeout: defaultPortWaitTimeoutStr, Locked: nil},
		},
	}
	for _, parsePortSpecTest := range parsePortSpecSuccessTests {
		t.Run(parsePortSpecTest.name, func(t *testing.T) {
			got, err := parsePortSpecStr(parsePortSpecTest.args.specStr)
			require.NoError(t, err, "Unexpected error occurred while testing")
			require.Equal(t, parsePortSpecTest.want, got)
		})
	}
}

func TestParsePortSpecstr_FailureCases(t *testing.T) {
	type args struct {
		specStr string
	}
	parsePortSpecFailureTests := []struct {
		name       string
		args       args
		errMessage string
	}{
		{
			name: "Failure while parsing, missing port number",
			args: args{
				specStr: "http:tcp",
			},
			errMessage: fmt.Sprintf("Error occurred while parsing port number '%v' in port spec '%v'", "tcp", "http:tcp"),
		},
		{
			name: "Failure while parsing, more than one delimiter ':'",
			args: args{
				specStr: "http:233:80",
			},
			errMessage: fmt.Sprintf("Error occurred while parsing port number '%v' in port spec '%v'", "233:80", "http:233:80"),
		},
		{
			name: "Failure while parsing, empty application protocol",
			args: args{
				specStr: ":3333/udp",
			},
			errMessage: fmt.Sprintf("Error occurred while parsing application protocol '%v' in port spec '%v'", "", ":3333/udp"),
		},
		{
			name: "Failure while parsing, extra delimeter(:) is present",
			args: args{
				specStr: "http:80/udp:",
			},
			errMessage: fmt.Sprintf("Error occurred while parsing transport protocol '%v' in port spec '%v'", "udp:", "http:80/udp:"),
		},
		{
			name: "Failure while parsing, port number is not a number",
			args: args{
				specStr: "http:abc/udp",
			},
			errMessage: fmt.Sprintf("Error occurred while parsing port number '%v' in port spec '%v'", "abc", "http:abc/udp"),
		},
	}
	for _, parsePortSpecTest := range parsePortSpecFailureTests {
		t.Run(parsePortSpecTest.name, func(t *testing.T) {
			_, err := parsePortSpecStr(parsePortSpecTest.args.specStr)
			require.NotNil(t, err, "Expected error, but received nil")
			require.ErrorContains(t, err, parsePortSpecTest.errMessage)
		})
	}
}

func TestParsePortSpecstr_EmptyIsError(t *testing.T) {
	_, err := parsePortSpecStr("")
	require.Error(t, err)
}

func TestParsePortSpecstr_AlphabeticalIsError(t *testing.T) {
	_, err := parsePortSpecStr("abd")
	require.Error(t, err)
}

func TestParsePortSpecstr_TooManyDelimsIsError(t *testing.T) {
	_, err := parsePortSpecStr("1234/tcp/udp")
	require.Error(t, err)
}

func TestParsePortSpecstr_DefaultTcpProtocol(t *testing.T) {
	portSpec, err := parsePortSpecStr("1234")
	require.NoError(t, err)
	require.Equal(t, uint32(1234), portSpec.GetNumber())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, portSpec.GetTransportProtocol())
}

func TestParsePortSpecstr_CustomProtocol(t *testing.T) {
	portSpec, err := parsePortSpecStr("1234/udp")
	require.NoError(t, err)
	require.Equal(t, uint32(1234), portSpec.GetNumber())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_UDP, portSpec.GetTransportProtocol())
}

func TestParsePortsStr_DuplicatePortsCauseError(t *testing.T) {
	_, err := parsePortsStr("http=80/tcp,http=8080")
	require.Error(t, err)
}

func TestParsePortsStr_EmptyPortIDCausesError(t *testing.T) {
	_, err := parsePortsStr("=80/tcp")
	require.Error(t, err)
}

func TestParsePortsStr_SuccessfulPortsString(t *testing.T) {
	ports, err := parsePortsStr("port1=8080,port2=2900/udp")
	require.NoError(t, err)
	require.Equal(t, 2, len(ports))

	port1Spec, found := ports["port1"]
	require.True(t, found)
	require.Equal(t, uint32(8080), port1Spec.GetNumber())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, port1Spec.GetTransportProtocol())

	port2Spec, found := ports["port2"]
	require.True(t, found)
	require.Equal(t, uint32(2900), port2Spec.GetNumber())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_UDP, port2Spec.GetTransportProtocol())
}

func TestParseEnvVarsStr_EqualSignInValueIsOkay(t *testing.T) {
	envvars, err := parseEnvVarsStr("VAR=thing=otherthing")
	require.NoError(t, err)
	require.Equal(t, 1, len(envvars))

	value, found := envvars["VAR"]
	require.True(t, found)
	require.Equal(t, "thing=otherthing", value)
}

func TestParseEnvVarsStr_MultipleVarsAreOkay(t *testing.T) {
	envvars, err := parseEnvVarsStr("VAR1=VALUE1,VAR2=VALUE2")
	require.NoError(t, err)
	require.Equal(t, 2, len(envvars))

	value1, found := envvars["VAR1"]
	require.True(t, found)
	require.Equal(t, "VALUE1", value1)

	value2, found := envvars["VAR2"]
	require.True(t, found)
	require.Equal(t, "VALUE2", value2)
}

func TestParseEnvVarsStr_DuplicateVarNamesError(t *testing.T) {
	_, err := parseEnvVarsStr("VAR1=VALUE1,VAR1=VALUE2")
	require.Error(t, err)
}

func TestParseEnvVarsStr_EmptyDeclarations(t *testing.T) {
	envvars, err := parseEnvVarsStr("VAR1=VALUE1,, ,  ,,")
	require.NoError(t, err)
	require.Equal(t, 1, len(envvars))
}

func TestParseFilesArtifactMountStr_ValidParse(t *testing.T) {
	artifactUuid1 := "1234"
	artifactUuid2 := "4567"
	mountpoint1 := "/dest1"
	mountpoint2 := "/dest2"

	result, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v:%v,%v:%v",
		mountpoint1,
		artifactUuid1,
		mountpoint2,
		artifactUuid2,
	))
	require.NoError(t, err)
	require.Equal(t, 2, len(result))

	parsedArtifactUuid1, found := result[mountpoint1]
	require.True(t, found)
	require.Equal(t, artifactUuid1, parsedArtifactUuid1)

	parsedArtifactUuid2, found := result[mountpoint2]
	require.True(t, found)
	require.Equal(t, artifactUuid2, parsedArtifactUuid2)
}

func TestParseFilesArtifactMountStr_EmptyDeclarationsAreSkipped(t *testing.T) {
	artifactUuid1 := "1234"
	artifactUuid2 := "4567"
	mountpoint1 := "/dest1"
	mountpoint2 := "/dest2"

	result, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v:%v,,,,,%v:%v",
		mountpoint1,
		artifactUuid1,
		mountpoint2,
		artifactUuid2,
	))
	require.NoError(t, err)
	require.Equal(t, 2, len(result))

	parsedArtifactUuid1, found := result[mountpoint1]
	require.True(t, found)
	require.Equal(t, artifactUuid1, parsedArtifactUuid1)

	parsedArtifactUuid2, found := result[mountpoint2]
	require.True(t, found)
	require.Equal(t, artifactUuid2, parsedArtifactUuid2)
}

func TestParseFilesArtifactMountStr_TooManyArtifactUuidMountpointDelimitersIsError(t *testing.T) {
	artifactUuid := services.FilesArtifactUUID("1234")
	mountpoint := "/dest"

	_, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v::%v",
		artifactUuid,
		mountpoint,
	))
	require.Error(t, err)
}

func TestParseFilesArtifactMountStr_TooFewArtifactUuidMountpointDelimitersIsError(t *testing.T) {
	artifactUuid := services.FilesArtifactUUID("1234")
	mountpoint := "/dest"

	_, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v%v",
		artifactUuid,
		mountpoint,
	))
	require.Error(t, err)
}

func TestParseFilesArtifactMountStr_DuplicateArtifactUuids(t *testing.T) {
	artifactUuid := services.FilesArtifactUUID("1234")
	mountpoint1 := "/dest1"
	mountpoint2 := "/dest2"

	_, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v:%v,%v:%v",
		artifactUuid,
		mountpoint1,
		artifactUuid,
		mountpoint2,
	))
	require.Error(t, err)
}

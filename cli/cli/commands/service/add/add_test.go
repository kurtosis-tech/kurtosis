package add

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

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
	require.Equal(t, uint16(1234), portSpec.GetNumber())
	require.Equal(t, services.PortProtocol_TCP, portSpec.GetProtocol())
}

func TestParsePortSpecstr_CustomProtocol(t *testing.T) {
	portSpec, err := parsePortSpecStr("1234/udp")
	require.NoError(t, err)
	require.Equal(t, uint16(1234), portSpec.GetNumber())
	require.Equal(t, services.PortProtocol_UDP, portSpec.GetProtocol())
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
	require.Equal(t, uint16(8080), port1Spec.GetNumber())
	require.Equal(t, services.PortProtocol_TCP, port1Spec.GetProtocol())

	port2Spec, found := ports["port2"]
	require.True(t, found)
	require.Equal(t, uint16(2900), port2Spec.GetNumber())
	require.Equal(t, services.PortProtocol_UDP, port2Spec.GetProtocol())
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
	artifactUuid1 := services.FilesArtifactUUID("1234")
	artifactUuid2 := services.FilesArtifactUUID("4567")
	mountpoint1 := "/dest1"
	mountpoint2 := "/dest2"

	result, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v:%v,%v:%v",
		artifactUuid1,
		mountpoint1,
		artifactUuid2,
		mountpoint2,
	))
	require.NoError(t, err)
	require.Equal(t, 2, len(result))

	parsedMountpoint1, found := result[artifactUuid1]
	require.True(t, found)
	require.Equal(t, mountpoint1, parsedMountpoint1)

	parsedMountpoint2, found := result[artifactUuid2]
	require.True(t, found)
	require.Equal(t, mountpoint2, parsedMountpoint2)
}

func TestParseFilesArtifactMountStr_EmptyDeclarationsAreSkipped(t *testing.T) {
	artifactUuid1 := services.FilesArtifactUUID("1234")
	artifactUuid2 := services.FilesArtifactUUID("4567")
	mountpoint1 := "/dest1"
	mountpoint2 := "/dest2"

	result, err := parseFilesArtifactMountsStr(fmt.Sprintf(
		"%v:%v,,,,,%v:%v",
		artifactUuid1,
		mountpoint1,
		artifactUuid2,
		mountpoint2,
	))
	require.NoError(t, err)
	require.Equal(t, 2, len(result))

	parsedMountpoint1, found := result[artifactUuid1]
	require.True(t, found)
	require.Equal(t, mountpoint1, parsedMountpoint1)

	parsedMountpoint2, found := result[artifactUuid2]
	require.True(t, found)
	require.Equal(t, mountpoint2, parsedMountpoint2)
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
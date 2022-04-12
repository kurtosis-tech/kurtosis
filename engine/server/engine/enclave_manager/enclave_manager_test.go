package enclave_manager

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	// !!!!!! BEFORE YOU UPDATE THIS CONSTANT TO FIX THE TEST, ADD A "BREAKING CHANGE" SECTION IN THE CHANGELOG !!!!!!
	// Explanation:
	//  * A breaking API change in Kurt Core must also yield a breaking API change in the engine server, since the
	//    engine server's KurtosisContext returns an EnclaveContext object from Core
	//  * However, we don't want to rely on Kurtosis dev memory to add a "Breaking Change" section to the engine
	//    server's changelog whenever a breaking Core API version change happens
	//  * Therefore, this constant must be manually updated to the X.Y version of the Core version you just
	//    bumped to, which will remind you to update the "Breaking Change" section of the changelog.
	expectedCoreMajorMinorVersion = "1.41"
	// !!!!!! BEFORE YOU UPDATE THIS CONSTANT TO FIX THE TEST, ADD A "BREAKING CHANGE" SECTION IN THE CHANGELOG !!!!!!
)

// This test ensures that when you bump to a Kurt Core version that has an API break, you're reminded to add a "Breaking Changes"
//  entry to the engine server's changelog as well (since a Kurt Core API break is an engine server API break)
func TestYouHaveBeenRemindedToAddABreakingChangelogEntryOnKurtCoreAPIBreak(t *testing.T) {
	actualKurtCoreSemver, err := semver.Parse(api_container_launcher.DefaultVersion)
	require.NoError(t, err, "An unexpected error occurred parsing Kurt Core version string '%v'", api_container_launcher.DefaultVersion)
	actualCoreMajorMinorVersion := fmt.Sprintf("%v.%v", actualKurtCoreSemver.Major, actualKurtCoreSemver.Minor)
	require.Equal(t, expectedCoreMajorMinorVersion, actualCoreMajorMinorVersion)
}

func TestOneToOneMappingBetweenObjAttrProtosAndDockerProtos(t *testing.T) {
	// Ensure all obj attr protos are used
	require.Equal(t, len(schema.AllowedProtocols), len(objAttrsSchemaPortProtosToDockerPortProtos))

	// Ensure all teh declared obj attr protos are valid
	for candidateObjAttrProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		_, found := schema.AllowedProtocols[candidateObjAttrProto]
		require.True(t, found, "Invalid object attribute schema proto '%v'", candidateObjAttrProto)
	}

	// Ensure no duplicate Docker protos, which is the best we can do since Docker doesn't expose an enum of all the protos they support
	seenDockerProtos := map[string]schema.PortProtocol{}
	for objAttrProto, dockerProto := range objAttrsSchemaPortProtosToDockerPortProtos {
		preexistingObjAttrProto, found := seenDockerProtos[dockerProto]
		require.False(t, found, "Docker proto '%v' is already in use by obj attr proto '%v'", dockerProto, preexistingObjAttrProto)
		seenDockerProtos[dockerProto] = objAttrProto
	}
}

func TestIsContainerRunningDeterminerCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, found := isContainerRunningDeterminer[containerStatus]
		require.True(t, found, "No is-container-running determination provided for container status '%v'", containerStatus.String())
	}
}
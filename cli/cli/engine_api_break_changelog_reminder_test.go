package main

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/stretchr/testify/require"
	"testing"
)

const (

	// !!!!!! BEFORE YOU UPDATE THIS CONSTANT TO FIX THE TEST, ADD A "BREAKING CHANGE" SECTION IN THE CHANGELOG !!!!!!
	// Explanation:
	//  * A breaking API change in the engine must also yield a breaking API change in the CLI, since a) users will
	//    have to update their modules and b) users will need to manually restart their engine
	//  * However, we don't want to rely on Kurtosis dev memory to add a "Breaking Change" section to the CLI
	//    changelog whenever an engine API break happens
	//  * Therefore, this constant must be manually updated to the X.Y version of the engine server version you just
	//    bumped to, which will remind you to update the "Breaking Change" section of the changelog.
	expectedEngineMajorMinorVersion = "1.21"
	// !!!!!! BEFORE YOU UPDATE THIS CONSTANT TO FIX THE TEST, ADD A "BREAKING CHANGE" SECTION IN THE CHANGELOG !!!!!!
)

// This test ensures that when you bump to a Kurt Core version that has an API break, you're reminded to add a "Breaking Changes"
//  entry to the engine server's changelog as well (since a Kurt Core API break is an engine server API break)
func TestYouHaveBeenRemindedToAddABreakingChangelogEntryOnKurtCoreAPIBreak(t *testing.T) {
	actualEngineSemver, err := semver.Parse(engine_server_launcher.KurtosisEngineVersion)
	require.NoError(t, err, "An unexpected error occurred parsing engine server version string '%v'", engine_server_launcher.KurtosisEngineVersion)
	actualEngineMajorMinorVersion := fmt.Sprintf("%v.%v", actualEngineSemver.Major, actualEngineSemver.Minor)
	require.Equal(t, expectedEngineMajorMinorVersion, actualEngineMajorMinorVersion)
}

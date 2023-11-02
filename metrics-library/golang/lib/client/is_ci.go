package client

import "os"

const (
	trueStr  = "true"
	falseStr = "false"
)

var ciEnvironmentVariables = []string{
	// Azure Pipelines
	"TF_BUILD",
	// Buildkite
	"BUILDKITE",
	// Circle
	"CIRCLECI",
	// Cirrus
	"CIRRUS_CI",
	// CodeBuild
	"CODEBUILD_ID",
	// Github
	"GITHUB_ACTIONS",
	// Gitlab
	"GITLAB_ACTIONS",
	// Heroku
	"HEROKU_TEST_RUN_ID",
	// Hudson & Jenkins
	"BUILD_ID",
	// TeamCity
	"TEAMCITY_VERSION",
	// Travis
	"TRAVIS",

	// Platform Agnostic Variables
	"CI",
}

// isCI Checks environment variables to tell if Kurtosis is running in CI
// This implements this blogpost https://adamj.eu/tech/2020/03/09/detect-if-your-tests-are-running-on-ci/
func isCI() string {
	for _, environmentVariable := range ciEnvironmentVariables {
		if _, found := os.LookupEnv(environmentVariable); found {
			return trueStr
		}
	}
	return falseStr
}

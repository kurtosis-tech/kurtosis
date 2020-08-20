package api

type StartServiceArgs struct {
	ImageName             string            `json:"imageName"`
	UsedPorts             []string          `json:"usedPorts"`
	StartCmd              []string          `json:"startCommand"`
	DockerEnvironmentVars map[string]string `json:"dockerEnvironmentVars"`
	TestVolumeMountFilepath string			`json:"testVolumeMountFilepath"`
}

type StartServiceResponse struct {
	IPAddress string 	`json:"ipAddress"`
}

type RegisterTestSuiteContainerArgs struct {
	// Docker container ID of the test suite container that's running
	ContainerId string

	// The testsuite container will be running a single test, and this lets the Kurtosis API know what the hard test
	//  timeout of that test will be (after which the Kurtosis API container will count the testsuite container as hung
	//  and report back to the initializer that the test didn't finish in the alloted time)
	TestTimeoutSeconds int
}

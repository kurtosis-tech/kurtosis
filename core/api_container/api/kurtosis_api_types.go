package api

type AddServiceArgs struct {
	// The user won't know the IP address of the service they're creating in advance, but they might need to use the IP
	//  in their start command. This string is the placeholder they used for it, which we'll substitute before launching
	//  the container
	IPPlaceholder		  string			`json:"ipPlaceholder"`

	ImageName             string            `json:"imageName"`
	UsedPorts             []string          `json:"usedPorts"`
	StartCmd              []string          `json:"startCommand"`
	DockerEnvironmentVars map[string]string `json:"dockerEnvironmentVars"`
	TestVolumeMountFilepath string			`json:"testVolumeMountFilepath"`
}

type AddServiceResponse struct {
	ServiceID string	`json:"serviceId"`
	IPAddress string 	`json:"ipAddress"`
}

type RemoveServiceArgs struct {
	ServiceID string 	`json:"serviceId"`
}

type RegisterTestExecutionArgs struct {
	// The testsuite container will be running a single test, and this lets the Kurtosis API know what the hard test
	//  timeout of that test will be (after which the Kurtosis API container will count the testsuite container as hung
	//  and report back to the initializer that the test didn't finish in the alloted time)
	TestTimeoutSeconds int
}

type RegisterTestSuiteArgs struct {
	TestNames []string
}

type ListTestsResponse struct {
	IsTestSuiteRegistered	bool
	Tests 					[]string
}
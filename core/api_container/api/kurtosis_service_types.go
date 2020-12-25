/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

type ServiceID string

type AddServiceArgs struct {
	// The ID that the service will be identified by going forward
	ServiceID		      string			`json:"serviceId"`

	// The partition that the new service should be added to
	// Empty or nil == default partition
	PartitionID			  string			`json:"partitionId"`

	ImageName             string            `json:"imageName"`

	// The user won't know the IP address of the service they're creating in advance, but they might need to use the IP
	//  in their start command. This string is the placeholder they used for it, which we'll substitute before launching
	//  the container
	IPPlaceholder		  string			`json:"ipPlaceholder"`

	// This is in Docker port specification syntax, e.g. "80" (default TCP) or "80/udp"
	// It might even support ranges (e.g. "90:100/tcp"), though this is untested as of 2020-12-08
	UsedPorts             []string          `json:"usedPorts"`

	StartCmd              []string          `json:"startCommand"`
	DockerEnvironmentVars map[string]string `json:"dockerEnvironmentVars"`
	TestVolumeMountFilepath string			`json:"testVolumeMountFilepath"`
}

type AddServiceResponse struct {
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


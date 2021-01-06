/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

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
	TestVolumeMountDirpath string			`json:"testVolumeMountDirpath"`

	// Artifacts containing compressed files which should be unzipped and then
	//  mounted at the given location for the service
	// The ID of the artifact will correspond to the artifacts declared in the test metadata
	FilesArtifactMounts map[string]string
}

type AddServiceResponse struct {
	IPAddress string 	`json:"ipAddress"`
}

type RemoveServiceArgs struct {
	ServiceID string 	`json:"serviceId"`
	ContainerStopTimeoutSeconds int `json:"containerStopTimeoutSeconds"`
}

type PartitionConnectionInfo struct {
	IsBlocked bool		`json:"isBlocked"`
}

type RepartitionArgs struct {
	// Mapping of partition ID -> "set" of service IDs
	PartitionServices map[string]map[string]bool	`json:"partitionServices"`

	// Mapping of partitionA -> partitionB -> partitionConnection details
	// We use this format because JSON doesn't allow object keys
	// This format allows for the same connection to be defined twice, but we'll do error-checking to catch it
	PartitionConnections map[string]map[string]PartitionConnectionInfo `json:"partitionConnections"`

	DefaultConnection PartitionConnectionInfo `json:"defaultConnection"`
}

type RegisterTestExecutionArgs struct {
	// The testsuite container will be running a single test, and this lets the Kurtosis API know what the hard test
	//  timeout of that test will be (after which the Kurtosis API container will count the testsuite container as hung
	//  and report back to the initializer that the test didn't finish in the alloted time)
	TestTimeoutSeconds int	`json:"testTimeoutSeconds"`
}

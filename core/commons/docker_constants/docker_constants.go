/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package docker_constants

const (
	// This is the magic domain name inside a container that Docker will give the host machine running Docker itself
	// This is available by default on Docker for Mac & Windows because they run in VMs, but needs to be specifically
	//  bound in Docker for Linux
	HostMachineDomainInsideContainer = "host.docker.internal"

	// HostGatewayName is the string value that Docker will replace by
	// the value of HostGatewayIP daemon config value
	HostGatewayName = "host-gateway"
)

/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_manager

const (
	NetAdmin  ContainerCapability = "NET_ADMIN"
	SysPtrace ContainerCapability = "SYS_PTRACE"
)

type ContainerCapability string

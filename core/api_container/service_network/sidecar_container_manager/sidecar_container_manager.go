/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

type SidecarContainerManager interface {
	CreateSidecarContainer(
	) error

	DestroySidecarContainer(
	) error
}
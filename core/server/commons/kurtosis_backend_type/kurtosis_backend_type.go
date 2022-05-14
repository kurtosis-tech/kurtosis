/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kurtosis_backend_type

//go:generate go run github.com/dmarkham/enumer -type=KurtosisBackendType -transform=lower
type KurtosisBackendType uint
const (
	// To add new values, just add a new version to the end WITHOUT WHITESPACE
	Docker KurtosisBackendType = iota
	Kubernetes
)
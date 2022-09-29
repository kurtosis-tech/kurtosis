/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

//go:generate go run github.com/dmarkham/enumer -trimprefix "KurtosisBackendType_" -type=KurtosisBackendType -transform=lower -json
type KurtosisBackendType uint
const (
	// To add new values, just add a new value to the end WITHOUT WHITESPACE
	KurtosisBackendType_Docker KurtosisBackendType = iota
	KurtosisBackendType_Kubernetes
)
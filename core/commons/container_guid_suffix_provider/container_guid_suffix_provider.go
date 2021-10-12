/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package container_guid_suffix_provider

import (
	"strconv"
	"time"
)

const (
	// TODO Change this to base 16 to be more compact??
	guidBase = 10
)

// Provides a unique suffix for container GUIDs, so that containers with the same ID (e.g. Lambda ID, service ID, etc.)
//  won't overlap
func GetContainerGUIDSuffix() string {
	now := time.Now()
	// TODO make this UnixNano to reduce risk of collisions???
	nowUnixSecs := now.Unix()
	return strconv.FormatInt(nowUnixSecs, guidBase)
}

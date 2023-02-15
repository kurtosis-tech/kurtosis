/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package current_time_str_provider

import (
	"strconv"
	"time"
)

const (
	// TODO Change this to base 16 to be more compact??
	guidBase = 10
)

// Provides the current time in string form, for use as a suffix to a container ID (e.g. service ID, module ID) that will
//  make it unique so it won't collide with other containers with the same ID
func GetCurrentTimeStr() string {
	now := time.Now()
	// TODO make this UnixNano to reduce risk of collisions???
	nowUnixSecs := now.Unix()
	return strconv.FormatInt(nowUnixSecs, guidBase)
}

/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package execution_ids

import (
	"fmt"
	"time"
)

const (
	kurtosisPrefix = "KT"

	// YYYY-MM-DDTHH.MM.SS.sss
	executionTimestampFormat = "2006-01-02T15.04.05.000"
)

func GetExecutionID() string {
	return fmt.Sprintf(
		"%v%v",
		kurtosisPrefix,
		time.Now().Format(executionTimestampFormat),
	)
}

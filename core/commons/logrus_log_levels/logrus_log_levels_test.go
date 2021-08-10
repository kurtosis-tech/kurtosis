/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logrus_log_levels

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAcceptableLogLevels(t *testing.T) {
	logLevelStrs := GetAcceptableLogLevelStrs()
	for _, str := range logLevelStrs {
		assert.NotEqual(t, "", str)
	}
	assert.Equal(t, len(logrus.AllLevels), len(logLevelStrs))
}

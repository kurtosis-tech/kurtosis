/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"gotest.tools/assert"
	"net"
	"testing"
)

func TestIpPlaceholderReplacement(t *testing.T) {
	ipPlaceholder := "FOOBAR"
	inputStartCmd := []string{
		// should not get replaced
		"foobar",
		// should get replaced
		ipPlaceholder,
		// should get double-replaced
		"aaa" + ipPlaceholder + ipPlaceholder + "bbb",
		// nothing should happen
		"",
	}

	inputEnvVars := map[string]string{
		// Key should NOT get replaced,
		ipPlaceholder: "nothing",

		// Value SHOULD get replaced
		"something": ipPlaceholder,
	}

	ipStr := "172.10.0.31"
	realIp := net.ParseIP(ipStr)

	outputStartCmd, outputEnvVars := replaceIpPlaceholderForDockerParams(
		ipPlaceholder,
		realIp,
		inputStartCmd,
		inputEnvVars)

	expectedStartCmd := []string{
		"foobar",
		ipStr,
		"aaa" + ipStr + ipStr + "bbb",
		"",
	}

	expectedEnvVars := map[string]string{
		ipPlaceholder: "nothing",
		"something": ipStr,
	}

	assert.DeepEqual(t, outputStartCmd, expectedStartCmd)
	assert.DeepEqual(t, outputEnvVars, expectedEnvVars)
}
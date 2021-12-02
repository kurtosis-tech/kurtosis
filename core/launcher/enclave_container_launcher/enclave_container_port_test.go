/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_container_launcher

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInvalidProtocol(t *testing.T) {
	_, err := NewEnclaveContainerPort(uint16(1234), "abcd")
	require.Error(t, err, "Expected an error on invalid protocol")
}

/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestValidUrls(t *testing.T) {
	for _, url := range urlsToValidateInTest {
		resp, err := http.Get(url)
		assert.NoError(t, err, "Got an unexpected error checking url '%v'", url)
		resp.Body.Close()
	}
}

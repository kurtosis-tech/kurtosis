/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_support_constants

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

const (
	safariUserAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	userAgentHeaderKey = "User-Agent"
)

func TestValidUrls(t *testing.T) {
	for _, url := range urlsToValidateInTest {
		jar, err := cookiejar.New(nil)
		if err != nil {
			assert.NoError(t, err, "Got an unexpected error creating the cookie jar")
		}
		// nolint: exhaustruct
		client := &http.Client{
			Jar: jar,
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "Got an unexpected error while creating a new GET request with URL '%v'", url)
		// Adding the User-Agent header because it's mandatory for sending a request to Twitter
		req.Header.Set(userAgentHeaderKey, safariUserAgent)
		resp, err := client.Do(req)
		assert.NoError(t, err, "Got an unexpected error checking url '%v'", url)
		assert.True(t, isValidReturnCode(resp.StatusCode), "URL '%v' returned unexpected status code: '%d'", url, resp.StatusCode)
		assert.NoError(t, err, "Got an unexpected error checking url '%v'", url)
		resp.Body.Close()
	}
}

func isValidReturnCode(code int) bool {
	return code >= 200 && code < 400
}

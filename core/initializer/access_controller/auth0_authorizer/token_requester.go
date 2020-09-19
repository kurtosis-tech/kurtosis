/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// Response from token endpoint: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#receive-tokens
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope string `json:"scope"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}


func requestAuthToken(queryParams map[string]string) (tokenResponse *TokenResponse, err error) {
	// Prepare request for token endpoint
	url := auth0UrlBase + auth0TokenPath
	payloadString := ""
	for variable, value := range queryParams {
		payloadString += fmt.Sprintf("&%s=%s", variable, value)
	}
	payload := strings.NewReader(payloadString)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Execute request
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = httpRetryMax
	// Set retryClient logger off, otherwise you get annoying logs every request. https://github.com/hashicorp/go-retryablehttp/issues/31
	retryClient.Logger = nil

	res, err := retryClient.StandardClient().Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to poll for valid token.")
	}
	defer res.Body.Close()
	// TODO TODO TODO make unauthorized response catching more specific to expected errors
	if res.StatusCode >= 400 && res.StatusCode <= 499 {
		/*
			If the user has not yet logged in and authorized the device,
			auth0 will return a 4xx response: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#token-responses
		*/
		return nil, nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to read response body.")
	}

	tokenResponse = new(TokenResponse)
	json.Unmarshal(body, &tokenResponse)
	logrus.Tracef("Response from polling token: %+v", tokenResponse)
	return tokenResponse, nil
}

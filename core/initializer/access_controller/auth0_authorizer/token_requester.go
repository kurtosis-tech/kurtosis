/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const (
	audience = "https://api.kurtosistech.com/login"
	auth0UrlBase = "https://dev-lswjao-7.us.auth0.com"
	auth0DeviceAuthPath = "/oauth/device/code"
	auth0TokenPath = "/oauth/token"

	contentTypeHeaderName = "content-type"


	clientIdQueryParamName = "client_id"
	grantTypeQueryParamName = "grant_type"
	audienceQueryParam = "audience"
)

// Response from token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope string `json:"scope"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}


func requestAuthToken(params map[string]string, headers map[string]string) (tokenResponse *TokenResponse, err error) {
	// Prepare request for token endpoint
	url := auth0UrlBase + auth0TokenPath
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse body parameters.")
	}
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to make HTTP request.")
	}
	for variable, value := range headers {
		req.Header.Add(variable, value)
	}
	logrus.Debugf("Request: %+v", req)

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
		logrus.Debugf("Received an error code: %v", res.StatusCode)
		logrus.Debugf("Full response: %+v", res)
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

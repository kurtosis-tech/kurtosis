/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	auth0TokenPath = "oauth/token"

	contentTypeHeaderName = "content-type"

	jsonHeaderType = "application/json"
	formContentType = "application/x-www-form-urlencoded"

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
	url := auth0_constants.Issuer + auth0TokenPath
	contentType := headers[contentTypeHeaderName]

	var paramReader io.Reader
	switch contentType {
	case "":
		return nil, stacktrace.NewError("Headers must have a content-type header.")

	case jsonHeaderType:
		requestBody, err := json.Marshal(params)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to serialize body parameters.")
		}
		paramReader = bytes.NewBuffer(requestBody)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to create byte form of request body.")
		}

	case formContentType:
		payloadString := ""
		for variable, value := range params {
			payloadString += fmt.Sprintf("&%s=%s", variable, value)
		}
		paramReader = strings.NewReader(payloadString)

	default:
		return nil, stacktrace.NewError("Unrecognized content type header: %s", contentType)
	}


	req, err := http.NewRequest("POST", url, paramReader)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to make HTTP request.")
	}

	for variable, value := range headers {
		req.Header.Add(variable, value)
	}
	logrus.Tracef("Request: %+v", req)

	// Execute request
	retryClient := getConstantBackoffRetryClient()
	res, err := retryClient.Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to poll for valid token.")
	}
	defer res.Body.Close()
	// TODO TODO TODO handle fatal error codes when Auth0 absolutely won't return a
	if res.StatusCode != 200 {
		logrus.Tracef("Received an error code: %v", res.StatusCode)
		logrus.Tracef("Full response: %+v", res)
		/*
			If the user has not yet logged in and authorized the device,
			auth0 will return a 4xx response: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#token-responses
		*/
		return nil, stacktrace.NewError("Expected 200 status code when requesting auth token but got %v", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to read response body.")
	}

	tokenResponse = new(TokenResponse)
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to deserialize response from token endpoint.")
	}
	logrus.Tracef("Response from polling token: %+v", tokenResponse)
	return tokenResponse, nil
}

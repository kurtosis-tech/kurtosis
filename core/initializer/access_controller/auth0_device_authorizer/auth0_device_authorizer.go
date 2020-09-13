/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_device_authorizer

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*
	Implements the Device Code authorization flow from auth0: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow
	At a high level:
		1. Request device code (Device Flow): Request a device code that the user can use to authorize the device.
		2. Request device activation (Device Flow): Request that the user authorize the device using their laptop or smartphone.
		3. Request tokens (Device Flow): Poll the token endpoint to request a token.
		4. Authorize user (Browser Flow): The user authorizes the device, so the device can receive tokens.
		5. Receive tokens (Device Flow): After the user successfully authorizes the device, receive tokens.
 */

const (
	RequiredScope = "execute:kurtosis-core"

	audience = "https://api.kurtosistech.com/login"
	auth0UrlBase = "https://dev-lswjao-7.us.auth0.com"
	auth0DeviceAuthPath = "/oauth/device/code"
	auth0TokenPath = "/oauth/token"
	httpRetryMax = 5
	// Client ID for the Auth0 application pertaining to local dev workflows. https://auth0.com/docs/flows/device-authorization-flow#device-flow
	localDevClientId = "ZkDXOzoc1AUZt3dAL5aJQxaPMmEClubl"
	pollTimeout = 5 * 60 * time.Second
	requestTokenPayloadStringBase = "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
)

// Response from device code endpoint: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#device-code-response
type DeviceCodeResponse struct {
	DeviceCode string `json:"device_code"`
	UserCode string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn int `json:"expires_in"`
	Interval int `json:"interval"`
}

// Response from token endpoint: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#receive-tokens
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope string `json:"scope"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}

/*
	Prompts the user to click on a URL in which they will input their credentials.
	They will also need to confirm that this device is the one they are logging in on and confirming.
	Auth0 is polled until it confirms that the user has successfully logged in and confirmed their device.
 */

func AuthorizeUserDevice() (*TokenResponse, error) {
	// Prepare to request device code.
	url := auth0UrlBase + auth0DeviceAuthPath
	payload := strings.NewReader(
		fmt.Sprintf("client_id=%s&scope=%s&audience=%s", localDevClientId, RequiredScope, audience))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Send request for device code.
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = httpRetryMax
	// Set retryClient logger off, otherwise you get annoying logs every request. https://github.com/hashicorp/go-retryablehttp/issues/31
	retryClient.Logger = nil

	res, err := retryClient.StandardClient().Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to request device authorization from auth provider.")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to read response body.")
	}

	// Parse response from device code endpoint
	var deviceCodeResponse = new(DeviceCodeResponse)
	json.Unmarshal(body, &deviceCodeResponse)

	// Prompt user to access authentication URL in browser in order to authenticate.
	logrus.Infof("Please login to use Kurtosis by going to: %s\n Your user code for this device is: %s", deviceCodeResponse.VerificationUriComplete, deviceCodeResponse.UserCode)

	// Poll for token while the user authenticates and confirms their device.
	tokenResponse, err := pollForToken(deviceCodeResponse.DeviceCode, deviceCodeResponse.Interval)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to poll for token.")
	}
	return tokenResponse, nil
}


// ========================== HELPER FUNCTIONS ============================

/*
	Repeatedly polls the request token endpoint from auth0 to check if the user
	has authenticated successfully. For more information: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#request-tokens
*/
func pollForToken(deviceCode string, interval int) (*TokenResponse, error) {
	// Set up a ticker to mark intervals for polling
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Poll token endpoint at intervals until timeout is hit.
	for {
		select {
		case <-time.After(pollTimeout):
			return nil, stacktrace.NewError("Timed out waiting for user to authorize device.")
		case t := <-ticker.C:
			logrus.Tracef("Polling for token at %s\n", t)
			tokenResponse, err := requestToken(deviceCode)
			if err != nil {
				// ignore errors while polling so temporary network blips don't break functionality
				continue
			}
			if tokenResponse != nil {
				return tokenResponse, nil
			}
		}
	}
}


/*
	Given a device code for the device running kurtosis, checks with auth0 if the user has logged in
	and confirmed this device is theirs. For more information: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#request-tokens
	If the user has not yet signed in to auth0, will return nil, nil.
*/
func requestToken(deviceCode string) (tokenResponse *TokenResponse, err error) {
	// Prepare request for token endpoint
	url := auth0UrlBase + auth0TokenPath
	payloadString := requestTokenPayloadStringBase
	payloadString += fmt.Sprintf("&device_code=%s&client_id=%s", deviceCode, localDevClientId)
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

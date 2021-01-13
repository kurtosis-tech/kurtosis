/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authenticators

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*
	Authenticates and authorizes a user running Kurtosis on their own device.
	User verifies their device, and authenticated with username and password

	Implements the Device Code authentication flow from auth0: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow
	At a high level:
		1. Request device code (Device Flow): Request a device code that the user can use to authorize the device.
		2. Request device activation (Device Flow): Request that the user authorize the device using their laptop or smartphone.
		3. Request tokens (Device Flow): Poll the token endpoint to request a token.
		4. User authentication (Browser Flow): The user authenticates and authorizes the device, so the device can receive tokens.
		5. Receive tokens (Device Flow): After the user successfully authorizes the device, receive tokens.
 */

const (
	auth0DeviceAuthPath          = "oauth/device/code"
	// Client ID for the Auth0 application pertaining to local dev workflows. https://auth0.com/docs/flows/device-authorization-flow#device-flow
	localDevClientId = "ZkDXOzoc1AUZt3dAL5aJQxaPMmEClubl"
	pollTimeout = 5 * 60 * time.Second
	deviceCodeQueryParamName = "device_code"
	deviceCodeGrantType = "urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
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

// Extracted as an interface so mocks can be written for testing
type DeviceCodeAuthenticator interface {
	AuthorizeDeviceAndAuthenticate() (string, error)
}

type StandardDeviceCodeAuthenticator struct{}

func NewStandardDeviceCodeAuthenticator() *StandardDeviceCodeAuthenticator {
	return &StandardDeviceCodeAuthenticator{}
}

/*
	Prompts the user to click on a URL in which they will input their credentials.
	They will also need to confirm that this device is the one they are logging in on and confirming.
	Auth0 is polled until it confirms that the user has successfully logged in and confirmed their device.
 */

func (authenticator StandardDeviceCodeAuthenticator) AuthorizeDeviceAndAuthenticate() (string, error) {
	logrus.Trace("Authorizing user device...")

	// Prepare to request device code.
	url := auth0_constants.Issuer + auth0DeviceAuthPath
	payloadContents := fmt.Sprintf(
		"client_id=%s&audience=%s",
		localDevClientId,
		auth0_constants.Audience)
	logrus.Debugf("Payload contents: %v", payloadContents)
	payload := strings.NewReader(payloadContents)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	// Send request for device code.
	logrus.Trace("Requesting device authorization...")
	retryClient := getConstantBackoffRetryClient()
	res, err := retryClient.Do(req)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to request device authorization from auth provider.")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", stacktrace.NewError("Expected 200 status code when requesting device code but got %v", res.StatusCode)
	}

	logrus.Trace("Reading device authorization response body...")
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to read response body.")
	}
	logrus.Debugf("Device authorization response body: %v", string(body))

	// Parse response from device code endpoint
	logrus.Trace("Parsing device authorization response JSON...")
	var deviceCodeResponse = new(DeviceCodeResponse)
	if err := json.Unmarshal(body, &deviceCodeResponse); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the device code response")
	}

	// Prompt user to access authentication URL in browser in order to authenticate.
	logrus.Infof("Please login to use Kurtosis by going to: %s\n Your user code for this device is: %s", deviceCodeResponse.VerificationUriComplete, deviceCodeResponse.UserCode)

	// Poll for token while the user authenticates and confirms their device.
	token, err := pollForToken(deviceCodeResponse.DeviceCode, deviceCodeResponse.Interval)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to poll for token.")
	}
	return token, nil
}


// ========================== HELPER FUNCTIONS ============================

/*
	Repeatedly polls the request token endpoint from auth0 to check if the user
	has authenticated successfully. For more information: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#request-tokens
*/
func pollForToken(deviceCode string, interval int) (string, error) {
	// Set up a ticker to mark intervals for polling
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// initialize device query map
	params := map[string]string{
		grantTypeQueryParamName:  deviceCodeGrantType,
		deviceCodeQueryParamName: deviceCode,
		clientIdQueryParamName:   localDevClientId,
	}

	deviceCodeHeaderParams := map[string]string{contentTypeHeaderName: formContentType}

	// Poll token endpoint at intervals until timeout is hit.
	for {
		select {
		case <-time.After(pollTimeout):
			return "", stacktrace.NewError("Timed out waiting for user to authorize device.")
		case t := <-ticker.C:
			logrus.Tracef("Polling for token at %s\n", t)
			tokenResponse, err := requestAuthToken(params, deviceCodeHeaderParams)
			if err != nil {
				// ignore errors while polling so temporary network blips don't break functionality
				continue
			}
			if tokenResponse != nil {
				return tokenResponse.AccessToken, nil
			}
		}
	}
}

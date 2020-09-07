/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	auth0UrlBase = "https://dev-lswjao-7.us.auth0.com"
	auth0DeviceAuthPath = "/oauth/device/code"
	auth0TokenPath = "/oauth/token"
	localDevClientId = "ZkDXOzoc1AUZt3dAL5aJQxaPMmEClubl"
	requiredScope = "execute:kurtosis-core"
	audience = "https://api.kurtosistech.com/login"
	pollTimeout = 60 * time.Second
)

type DeviceCodeResponse struct {
	DeviceCode string `json:"device_code"`
	UserCode string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	VerificationUriComplete string `json:"verification_uri_complete"`
	ExpiresIn string `json:"expires_in"`
	Interval int `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope string `json:"scope"`
	ExpiresIn int `json:"expires_in"`
	TokenType string `json:"token_type"`
}

func authorizeUserDevice() (*TokenResponse, error) {
	var deviceCodeResponse = new(DeviceCodeResponse)
	url := auth0UrlBase + auth0DeviceAuthPath
	payload := strings.NewReader(
		fmt.Sprintf("client_id=%s&scope=%s&audience=%s", localDevClientId, requiredScope, audience))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &deviceCodeResponse)
	logrus.Debugf("Device Code: %+v\n", deviceCodeResponse)
	logrus.Infof("Please login to use Kurtosis by going to: %s\n Your user code for this device is: %s", deviceCodeResponse.VerificationUriComplete, deviceCodeResponse.UserCode)
	tokenResponse, err := pollForToken(deviceCodeResponse.DeviceCode, deviceCodeResponse.Interval)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to poll for token.")
	}
	return tokenResponse, nil
}

/*
	Requests an access token for the user from auth0.
	If the user has not yet signed in to auth0, will return nil, nil.
*/
func requestToken(deviceCode string) (tokenResponse *TokenResponse, err error) {
	tokenResponse = new(TokenResponse)
	url := auth0UrlBase + auth0TokenPath
	payloadString := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
	payloadString += fmt.Sprintf("&device_code=%s&client_id=%s", deviceCode, localDevClientId)
	payload := strings.NewReader(payloadString)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to poll for valid token.")
	}
	// TODO TODO TODO make unauthorized response catching more specific to expected errors
	if res.StatusCode >= 400 && res.StatusCode <= 499 {
		/*
			If the user has not yet logged in and authorized the device,
			auth0 will return a 4xx response: https://auth0.com/docs/flows/call-your-api-using-the-device-authorization-flow#token-responses
		*/
		return nil, nil
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &tokenResponse)
	logrus.Tracef("Response from polling token: %+v", tokenResponse)
	return tokenResponse, nil
}

func pollForToken(deviceCode string, interval int) (*TokenResponse, error) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(pollTimeout)
		done <- true
	}()
	for {
		select {
		case <-done:
			return nil, stacktrace.NewError("Timed out waiting for user to authorize device.")
		case t := <-ticker.C:
			logrus.Tracef("Polling for token at %s\n", t)
			tokenResponse, err := requestToken(deviceCode)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to request authentication token.")
			} else if tokenResponse != nil {
				return tokenResponse, nil
			}
		}
	}
}

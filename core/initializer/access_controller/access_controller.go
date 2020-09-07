/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	freeTrialLicense = "free-trial"
	expiredTrialLicense = "expired-trial"
	paidLicense = "paid-license"
	auth0UrlBase = "https://dev-lswjao-7.us.auth0.com"
	auth0DeviceAuthPath = "/oauth/device/code"
	auth0TokenPath = "/oauth/token"
	auth0AuthorizePath = "/authorize"
	localDevClientId = "ZkDXOzoc1AUZt3dAL5aJQxaPMmEClubl"
	callbackUri = "https://api.kurtosistech.com/success"
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


/*
   Implements the PKCE authentication workflow:
   https://auth0.com/docs/flows/call-your-api-using-the-authorization-code-flow-with-pkce
 */

func generateCodeVerifierAndChallenge() (verifier string, challenge string, err error) {
	randomSecureKey := [32]byte{}
	_, err = rand.Read(randomSecureKey[:])
	if err != nil {
		return "", "", stacktrace.Propagate(err, "")
	}
	challengeBytes := sha256.Sum256(randomSecureKey[:])
	verifier = b64.URLEncoding.EncodeToString(randomSecureKey[:])
	challenge = b64.URLEncoding.EncodeToString(challengeBytes[:])
	return verifier, challenge,nil
}

func authorizeUserDevice() error {
	var deviceCode = new(DeviceCodeResponse)
	url := auth0UrlBase + auth0DeviceAuthPath
	payload := strings.NewReader(
		fmt.Sprintf("client_id=%s&scope=%s&audience=%s", localDevClientId, requiredScope, audience))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	json.Unmarshal(body, &deviceCode)
	fmt.Printf("Device Code: %+v\n", deviceCode)
	err := pollForToken(deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to poll for token.")
	}
	return nil
}

func requestToken(deviceCode string) error {
	url := auth0UrlBase + auth0TokenPath
	payloadString := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code"
	payloadString += fmt.Sprintf("&device_code=%s&client_id=%s", deviceCode, localDevClientId)
	payload := strings.NewReader(payloadString)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to poll for valid token.")
	}
	fmt.Println(res)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
	return nil
}

func pollForToken(deviceCode string, interval int) error {
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
			fmt.Println("Timed out waiting for user to authorize device.")
			return nil
		case t := <-ticker.C:
			fmt.Printf("Grab at %s\n", t)
			requestToken(deviceCode)
		}
	}
	return nil
}

func authorizeUserPKCE() error {
	url := auth0UrlBase + auth0AuthorizePath
	_, challenge, err := generateCodeVerifierAndChallenge()
	if err != nil {
		return err
	}
	payloadString := fmt.Sprintf("?client_id=%s", localDevClientId)
	payloadString += fmt.Sprintf("&response_type=code")
	payloadString += fmt.Sprintf("&code_challenge_method=S256")
	payloadString += fmt.Sprintf("&code_challenge=%s", challenge)
	payloadString += fmt.Sprintf("&redirect_uri=%s", callbackUri)
	payloadString += fmt.Sprintf("&scope=%s", requiredScope)
	payloadString += fmt.Sprintf("&audience=%s", audience)
	payload := strings.NewReader(payloadString)
	fmt.Println(payloadString)

	req, _ := http.NewRequest("GET", url, payload)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(req)
	fmt.Println(res)
	fmt.Println(string(body))
	return nil
}

/*
	Verifies that the license is a credential for a registered user,
	Returns true if the license associated with the user has ever
	been registered in Kurtosis, even if that user has no permissions.
 */
func AuthenticateLicense(license string) (bool, error) {
	err := authorizeUserDevice()
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	switch license {
	case freeTrialLicense, paidLicense, expiredTrialLicense:
		return true, nil
	default:
		return false, nil
	}
}

/*
	Verifies that the user associated with the license is authorized
	to run kurtosis (either has a free trial or has purchased a full license).
*/
func AuthorizeLicense(license string) (bool, error) {
	switch license {
	case freeTrialLicense, paidLicense:
		return true, nil
	case expiredTrialLicense:
		return false, nil
	default:
		return false, stacktrace.NewError("Failed to authorize license.")
	}
}

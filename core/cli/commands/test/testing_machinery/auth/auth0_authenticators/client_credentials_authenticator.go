/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package auth0_authenticators

import (
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/auth0_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
	Used for machine-to-machine auth. This should be used for Kurtosis running in CI jobs.
	Runs the client credentials oAuth workflow:https://tools.ietf.org/html/rfc6749#section-4.4
	Implemented using the auth0 implementation of this flow: https://auth0.com/docs/flows/call-your-api-using-the-client-credentials-flow
 */

const (
	clientCredentialGrantType = "client_credentials"
	clientSecretQueryParamName = "client_secret"
)

// Extracted as an interface so mocks can be written for testing
type ClientCredentialsAuthenticator interface{
	AuthenticateClientCredentials(clientId string, clientSecret string) (string, error)
}

type StandardClientCredentialsAuthenticator struct{}

func NewStandardClientCredentialsAuthenticator() *StandardClientCredentialsAuthenticator {
	return &StandardClientCredentialsAuthenticator{}
}

func (authenticator StandardClientCredentialsAuthenticator) AuthenticateClientCredentials(clientId string, clientSecret string) (string, error) {
	params := map[string]string{
		clientIdQueryParamName:     clientId,
		clientSecretQueryParamName: clientSecret,
		grantTypeQueryParamName:    clientCredentialGrantType,
		audienceQueryParam:         auth0_constants.Audience,
	}
	headers := map[string]string{
		contentTypeHeaderName: jsonHeaderType,
	}

	tokenResponse, err := requestAuthToken(params, headers)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to get token response for client credential auth flow.")
	}
	logrus.Tracef("Token response: %+v", tokenResponse)

	return tokenResponse.AccessToken, nil
}

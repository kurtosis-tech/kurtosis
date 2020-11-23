/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizers

import (
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
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
type ClientCredentialsAuthorizer interface{
	AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error)
}

type StandardClientCredentialsAuthorizer struct{}

func NewStandardClientCredentialsAuthorizer() *StandardClientCredentialsAuthorizer {
	return &StandardClientCredentialsAuthorizer{}
}

func (authorizer StandardClientCredentialsAuthorizer) AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error) {
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
		return nil, stacktrace.Propagate(err, "Failed to get token response for client credential authorization flow.")
	}
	logrus.Tracef("Token response: %+v", tokenResponse)

	return tokenResponse, nil
}

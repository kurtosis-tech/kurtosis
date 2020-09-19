/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
	Implements the client credentials auth0 flow https://auth0.com/docs/flows/call-your-api-using-the-client-credentials-flow
 */

const (
	clientCredentialGrantType = "client_credentials"
	clientSecretQueryParamName = "client_secret"
	jsonHeaderType = "application/json"
)

func AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error) {
	params := map[string]string{
		clientIdQueryParamName: clientId,
		clientSecretQueryParamName: clientSecret,
		grantTypeQueryParamName: clientCredentialGrantType,
		audienceQueryParam: audience,
	}
	headers := map[string]string{
		contentTypeHeaderName: jsonHeaderType,
	}
	tokenResponse, err := requestAuthToken(params, headers)
	logrus.Debugf("Token response: %+v", tokenResponse)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get token response for client credential authorization flow.")
	}
	return tokenResponse, nil
}

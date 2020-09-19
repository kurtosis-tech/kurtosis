/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import "github.com/palantir/stacktrace"

/*
	Implements the client credentials auth0 flow https://auth0.com/docs/flows/call-your-api-using-the-client-credentials-flow
 */

const (
	clientCredentialGrantType = "client-credentials"
	clientSecretQueryParamName = "client_secret"
)

func AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error) {
	queryParams := map[string]string{
		clientIdQueryParamName: clientId,
		clientSecretQueryParamName: clientSecret,
		grantTypeQueryParamName: clientCredentialGrantType,
		audienceQueryParam: audience,
	}
	tokenResponse, err := requestAuthToken(queryParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get token response for client credential authorization flow.")
	}
	return tokenResponse, nil
}

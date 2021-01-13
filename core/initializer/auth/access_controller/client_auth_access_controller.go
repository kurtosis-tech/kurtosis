/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authenticators"
	"github.com/palantir/stacktrace"
)

type ClientAuthAccessController struct {
	// Mapping of key_id -> pem_encoded_pubkey_cert for validating tokens
	tokenValidationPubKeys map[string]string
	clientCredsAuthorizer  auth0_authenticators.ClientCredentialsAuthenticator
	clientId               string
	clientSecret           string
}

func NewClientAuthAccessController(
		tokenValidationPubKeys map[string]string,
		clientCredsAuthorizer auth0_authenticators.ClientCredentialsAuthenticator,
		clientId string,
		clientSecret string) *ClientAuthAccessController {
	return &ClientAuthAccessController{
		tokenValidationPubKeys: tokenValidationPubKeys,
		clientCredsAuthorizer: clientCredsAuthorizer,
		clientId: clientId,
		clientSecret: clientSecret,
	}
}

/*
This workflow is for authenticating Kurtosis tests running in CI (no device or username).
	See also: https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
*/
func (accessController ClientAuthAccessController) Authenticate() (*permissions.Permissions, error) {
	token, err := accessController.clientCredsAuthorizer.AuthenticateClientCredentials(
		accessController.clientId,
		accessController.clientSecret)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred authenticating with the client ID & secret")
	}

	claims, err := parseTokenClaims(accessController.tokenValidationPubKeys, token)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	perms := parsePermissionsFromClaims(claims)
	return perms, nil
}

/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authorizers"
	"github.com/palantir/stacktrace"
)

type ClientAuthAccessController struct {
	// Mapping of key_id -> base64_encoded_pubkey for validating tokens
	tokenValidationPubKeys map[string]string
	clientCredsAuthorizer auth0_authorizers.ClientCredentialsAuthorizer
	clientId string
	clientSecret string
}

func NewClientAuthAccessController(
		tokenValidationPubKeys map[string]string,
		clientCredsAuthorizer auth0_authorizers.ClientCredentialsAuthorizer,
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
This workflow is for authenticating and authorizing Kurtosis tests running in CI (no device or username).
	See also: https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
*/
func (accessController ClientAuthAccessController) Authorize() error {
	tokenResponse, err := accessController.clientCredsAuthorizer.AuthorizeClientCredentials(
		accessController.clientId,
		accessController.clientSecret)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred authenticating with the client ID & secret")
	}

	claims, err := parseTokenClaims(accessController.tokenValidationPubKeys, tokenResponse.AccessToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	if err := verifyExecutionPerms(claims); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying execution permissions")
	}
	return nil
}

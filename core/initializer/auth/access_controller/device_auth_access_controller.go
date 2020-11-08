/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authorizers"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/encrypted_session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type DeviceAuthAccessController struct {
	// Mapping of key_id -> base64_encoded_pubkey for validating tokens
	tokenValidationPubKeys map[string]string
	sessionCache encrypted_session_cache.EncryptedSessionCache
	deviceAuthorizer auth0_authorizers.DeviceAuthorizer
}

func NewDeviceAuthAccessController(
		tokenValidationPubKeys map[string]string,
		sessionCache encrypted_session_cache.EncryptedSessionCache,
		deviceAuthorizer auth0_authorizers.DeviceAuthorizer) *DeviceAuthAccessController {
	return &DeviceAuthAccessController{
		tokenValidationPubKeys: tokenValidationPubKeys,
		sessionCache: sessionCache,
		deviceAuthorizer: deviceAuthorizer,
	}
}

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them for their username and password

Returns:
	An error if and only if an irrecoverable login error occurred
*/
func (accessController DeviceAuthAccessController) Authorize() error {
	/*
		NOTE: As of 2020-10-24, we actually don't strictly *need* to encrypt anything on disk because we hardcode the
		 Auth0 public keys used for verifying tokens so unless the user cracks Auth0 and gets the private key, there's
		 no way for a user to forge a token.

		However, this hardcode-public-keys approach becomes much harder if we start doing private key rotation (which would
		 be good security hygiene) because:
		  a) now our code needs to dynamically discover what public keys it should use and
		  b) Kurtosis needs to work even if the developer is offline
		The offline requirement is the real kicker, because it means we need to write the public keys to the developer's local
		 machine and somehow protect it from tampering. This likely means encrypting the data, which means having an encryption
		 key in the code, which would shift the weakpoint to someone decompiling kurtosis-core and discovering the encryption
		 key there.
	*/
	tokenStr, err := getTokenStr(accessController.deviceAuthorizer, accessController.sessionCache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	logrus.Debugf("Token str (before expiry check): %v", tokenStr)

	claims, err := parseAndCheckTokenClaims(tokenStr, accessController.deviceAuthorizer, accessController.sessionCache)
	if err != nil {
		return stacktrace.Propagate(err, "An unrecoverable error occurred checking the token expiration")
	}

	if err := verifyExecutionPerms(claims); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying execution permissions")
	}
	return nil
}


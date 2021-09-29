/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package access_controller

import (
	permissions2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/access_controller/permissions"
	auth0_constants2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/auth0_constants"
	test_mocks2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/test_mocks"
	"testing"
)

func Test_UnableToAuth(t *testing.T) {
	errorThrowingAuthorizer := test_mocks2.NewMockClientCredentialsAuthorizer(true, "")
	accessController := NewClientAuthAccessController(test_mocks2.TestAuth0PublicKeys, errorThrowingAuthorizer, "id", "secret")

	_, err := accessController.Authenticate()
	if err == nil {
		t.Fatal("Expected an error due to an error with the authorizer, but access was permitted")
	}
}

func Test_TokenSignedByInvalidKey(t *testing.T) {
	randomPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	token, err := test_mocks2.CreateTestToken(
		randomPrivateKey,
		auth0_constants2.Audience,
		auth0_constants2.Issuer,
		3600,
		[]string{permissions2.UnlimitedTestExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token signed by the unknown private key: %v", err)
	}
	clientCredsAuthorizer := test_mocks2.NewMockClientCredentialsAuthorizer(false, token)

	// These public keys don't match the private key we've signed the token with; we expect a rejection
	accessController := NewClientAuthAccessController(
		test_mocks2.TestAuth0PublicKeys,
		clientCredsAuthorizer,
		"id",
		"secret",
	)
	_, err = accessController.Authenticate()
	if err == nil {
		t.Fatal("Expected an error due to a token signed by a key we don't recognize, but no error was thrown")
	}
}


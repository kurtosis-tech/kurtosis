/* * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_token_claims"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/session_cache"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/test_mocks"
	"github.com/palantir/stacktrace"
	"testing"
	"time"
)

const (
	testAuth0KeyId = "test-key-id"
	testAuth0PrivateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

)

var testAuth0PublicKeys = map[string]string{
	testAuth0KeyId: `-----BEGIN CERTIFICATE-----
MIIDsjCCAhoCCQDHUsSd2AcX7zANBgkqhkiG9w0BAQsFADAbMRkwFwYDVQQDDBBr
dXJ0b3Npcy10ZXN0aW5nMB4XDTIwMTEyMzAxNTc0NloXDTI1MTEyMjAxNTc0Nlow
GzEZMBcGA1UEAwwQa3VydG9zaXMtdGVzdGluZzCCAaIwDQYJKoZIhvcNAQEBBQAD
ggGPADCCAYoCggGBALkRXTuEca4hNvF2j0NpbLky6KQ//wuNmXLBsd+Bv7ihybf7
b1vEzGDjmy+DHc7CDSTgjyTrHPwRGBpkgBZH3+1GJJvTI+jtYS/IxwbNdOC25O7F
YRAgbj9i1vPR+V8FUNUkbzDqDpIYJDVHGnrr0PPZ9kYOUrE2pDh4MGfc0dkAZ06S
dNAuG3E1EN/G4+U4Hz/fGqBS3akRrisd2GL7Rj5h2vI+MY+TSLlfXCBbxANMEOqe
YaT34pmccWS5f7uTOeBePMgJHZM29bhy+s3TTCwjj7099SqseAMS7kiugIS0xYwO
2m4hTsl4FPtt0tbE9bWGis3RMfq0tiG4ErfMBpKakZaUj0SYYDwbOcwHTk7q9P6K
/KuTg1/leHlyTfnraCzT20jxU0CwiJTxVDWycdxc9b9+b6sN6NGrHFapdqgn4AvJ
vx58/XfKxl4Ereb6u7C0LOgRnpGtHQyqTHalrur8Ia7UL45oukbRIRXT5zqoCoQ3
w3S/IKzQkMDS1MXI6QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBgQCpVVi/AOaqTPk5
fMMgSljaYjoZ+XSflOl7Mtpfs465ZX6kuhdgGdizC9cpKW3DrqXKCwF/eGYanWmY
RtrpJmLI0ieKFIsvGRvuWnyEE2heAIBUjJgcC3m501hnoiKq2XuVhDSxcfif8CZ/
4WKYvbH+qigQt8xsrCBrpWlbApABPB4QBmdWdG2f/gca3pVKhUHzx1XDezTQ0sl/
pE1789TEKqY7Qm4QPjF1V3zk2WBfG3BLTqlHWynrQjo3oQcZdrk6WTWabPPnsUQr
ltkVRKZs5W31AylLPBpmw1cJ3+rsuB69j6OHnZ24mNKMejtQcrrEuVg5LcgZW6fO
fhnzy1v+7TVZFSr4yfHHn9Vb/OJn5+494TlZ8noKAc1C6+lLvam6+NOPbo0Dbsv/
r2qIAsNdjFbaK1RFC6eTxn5UICN43rlZf7SuPpcJMAIgKqe5zv8fVQ9Temz2qXQu
ba2/Jl+DFf+wRk0oYEt+Gw7KF6H4yS5MOmLyqSPijXdEdej4u68=
-----END CERTIFICATE-----`,
}

func Test_NoSavedSession_UnreachableAuth0(t *testing.T) {
	errorThrowingSessionCache := test_mocks.NewMockSessionCache(true, true, nil)
	errorThrowingDeviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(true, "abcd1234")

	// There's no session to load and an error reaching Auth0, so should throw an error
	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, errorThrowingSessionCache, errorThrowingDeviceAuthorizer)
	if err := accessController.Authorize(); err == nil {
		t.Fatal("Expected an error if no session cache could be loaded and Auth0 authorization failed, but no error was thrown")
	}
}

func Test_NoSavedSession_ReachableAuth0_UnparseableToken(t *testing.T) {
	loadingErrorSessionCache := test_mocks.NewMockSessionCache(false, true, nil)
	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, "abcd1234")

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, loadingErrorSessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err == nil {
		t.Fatal("Expected an error due to receiving an invalid token from Auth0, but no error was thrown")
	}
}

func Test_NoSavedSession_ReachableAuth0_ParseableToken_InvalidKey(t *testing.T) {
	loadingErrorSessionCache := test_mocks.NewMockSessionCache(false, true, nil)

	randomPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIG4gIBAAKCAYEAxDzdBUtZx5xnvwktFZjQsR+FbWXE0FFeiFJJRSBzRo/TK+jF
I6485iDolg/3/D5109lZXK1qCaZSzT1eiRPPbyZ2BVu8B3OZwjmspuVRkJyf4BMj
jMlOMuQNCKg2GgGoXHAfYzpVFkdpQKwIz2KqSZCzwtZSwaWjcgav4F1GLkO5YjPd
sHpPWr22QCgA8QmoJeEvGkYdTFLrolgiF6v4luHdUx6UP78YpN4Bz365sJf7bamI
HJcERJebJHnMbPygdH1FrnWHfnpFYuy1qLugZnSn7sUaFt2F7g3Qea2V+kTHFWa2
E8159KwyUnD7K7nVVPOzgqVhElUAq/bFvBiiqnzSb0KxvX4CNalaqO8eaSPY5OHa
P7iGs8TZfQ0WeDakd4mSn8C7RlQWPpreCiqSRQpNmkHdOFnVtgkTwYbRuyPI2WCo
RgKNdhrwrMy8eDqE4Chg7YVDgPHWCvKeXesw/ivCRpv8mGrRsiEhzcKom0q+hvGW
4azFKA/0Y7NXxCRZAgMBAAECggGAeJMdCr+9rlR/unWc6gQ3Vl2T0iARyh31A7Xr
pznFGroMepJPbxkD+jKGNo4hRS/rnfuSWMuEt+EmR01J5NfzQMxU//3ZjoqNEzX0
y6djcoOKCFg6I6sdDU/qYkNY2qniFMofvwx8c2/1T/NkhmiNUR5EFZcyyiFISCur
rSQilxKtuZU8xc6hK5Qdg1YRHglc88lk87PZFKhueBXG/NyBpdOp9gR3+qMkNUjj
u+aVmdgIeN0N8OMpkCvG+6aK4PxnVO51z6c4axVVCZibHzLs1L6LXqmNqjNmBq79
gFL9WBAlVU4V4fFBmbXX+KuJls+7cj+KgMfX7vW+LwGWMKr2jstakb3/qdXAgOEw
GHYHxNN+dwkoq4iafyR8gBD/UIqo6DolETP08Y79Dg7i2sH71tIg2r2ggRYmeP/d
9WVRnmgB5GxSxy3GjlAmOq3Hz8qtCLFtFuRQqrtBRvAx763C1hb/RWRFoI0RgLpd
rnuc5RihCugW5AY9DVGhE/+IdGDBAoHBAP8QTzqeaWhT1Iq3S8uS+YgO2wmrgFby
HFF7Oq3/fgzcI6+YZCiLnYTUHct002R7bzjnsUO1dRH4Bf3yHqc4u5WhpZ2e2IV1
+4mpgGWqP0Hwq93yNvi2rWyAJrtFuZgEOp2tbPDRinlN+ZK7C0rk21nsg93HVGvW
Sb7tMZnkI3AHNj7p6Myd13RYbkb9QXgaARRj4Rb02Cdyt5cjpvJisRTkL6OwgneD
8rbWiUm5BQgtDhA3lqM0PN2saBCuMNitJQKBwQDE9UYCI03hK0wOAdz1qPGvTlP+
rSQjpdXttkmC2LPxsmaWEzbJD1XYL3IXob0cjT3kil0hF+SIYCaRB6VQBD8bjrMU
kTMdF6aBeGiBblINKicjA0V2jXHIDOhvdBwyziKhNXKjNGv9zF7BnfRn0ISwlx1k
6J++nMaWAPXHSYflfAiYFbatv915G/eivWJZZQRYwloZO08nehv4xUWxd3IhHIMF
KEkYUJ6ontqtod1YpMIDXuAMpaM/sP3U76k/RiUCgcAXxpBsGWof9HiCebWSA2BJ
Q4E9dIQhFq53FfKRV5iLYFXfP2hOszz6rb8dQQWXfz4N4uMOObLw+tqsIk6jLdGm
kAvdFnp+blIFMgyq7WS6I9IRfUuMgZLG42c427YCKprAKfNWu2GaDx+tgsv5rj2Y
M0jTeoovBymWp4uRGcgH4FQ5JxqxQCFeUgPtkBvzMxFYsjrAJhCkFLhyWTttqq9x
EBg0vPZcZ6tPSc5AVgPXEEQYVOYwzmTCERkePO6GtBUCgcBZvPcc2kENqtCIQUkP
lN4pZaLXksO4ikKigD+WIm46XXJoRnDbwuT2DwgIxGSJscDVdEViYqR5jnWD9tvX
TVgDkkz9vfpv8uqmatoSvtUbsm0Kgt8PWPrSjy8IOPrwGwOkN9n3ilb52DgEN5e4
BUWvv+pgo6zFCGFizyUsm9ATOyQfRyVonNan65o0x90bpe8JEeRDQsaZ0gUUn61V
YnrZo0f+/Y/wSCtB4L76BZn4XXkYWA31NTLgPiAo+NlAPxECgcB+conl67OsrJji
JsVZ/EN7zZyLMTKluWVU566i2AdDNtrC6RVGAJl5n2M5yLhjgUCB8QolK2MWzNle
Z/VixDmRS8ju4PLRe4a8hIV+TFnZFqDaGSdUG36ygu/mi+dFjp8FmO5uXWLlPvrM
rxDITHSZGsHhPFMD1p2F7Jp8Hm1ja5flaDP6IdjLtqj1RfMmYcuQ0rkBgC+PN14Z
P++/eqe7bHhyQg2uH7nv/kNv0GX02qUtk8zZqMOyE6fW3wg1e/k=
-----END RSA PRIVATE KEY-----`

	token, err := createTestToken(
		randomPrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token signed by the unknown private key: %v", err)
	}
	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, token)

	// These public keys don't match the private key we've signed the token with; we expect a rejection
	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, loadingErrorSessionCache, deviceAuthorizer)
	err = accessController.Authorize()
	if err == nil {
		t.Fatal("Expected an error due to a token signed by a key we don't recognize, but no error was thrown")
	}
	savedSessions := loadingErrorSessionCache.GetSavedSessions()
	if len(savedSessions) != 1 {
		t.Fatalf("Expected one session to be saved (upon receiving the new token) but got %v", len(savedSessions))
	}
	firstSavedSession := savedSessions[0]
	if firstSavedSession.Token != token {
		t.Fatalf("Expected token '%v' to be saved but got '%v'", token, firstSavedSession.Token)
	}
}

func Test_NoSavedSession_ReachableAuth0_ParseableToken_NoPerms(t *testing.T) {
	loadingErrorSessionCache := test_mocks.NewMockSessionCache(false, true, nil)

	token, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token without any perms: %v", err)
	}
	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, token)

	// These public keys don't match the private key we've signed the token with; we expect a rejection
	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, loadingErrorSessionCache, deviceAuthorizer)
	err = accessController.Authorize()
	if err == nil {
		t.Fatal("Expected an error due to a token without execution perms, but no error was thrown")
	}
	savedSessions := loadingErrorSessionCache.GetSavedSessions()
	if len(savedSessions) != 1 {
		t.Fatalf("Expected one session to be saved (upon receiving the new token) but got %v", len(savedSessions))
	}
	firstSavedSession := savedSessions[0]
	if firstSavedSession.Token != token {
		t.Fatalf("Expected token '%v' to be saved but got '%v'", token, firstSavedSession.Token)
	}
}

func Test_NoSavedSession_ReachableAuth0_ParseableToken_Valid(t *testing.T) {
	loadingErrorSessionCache := test_mocks.NewMockSessionCache(false, true, nil)

	token, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token: %v", err)
	}
	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, token)

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, loadingErrorSessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err != nil {
		t.Fatalf("We expected a successful authorization, but an error was thrown: %v", err)
	}
	savedSessions := loadingErrorSessionCache.GetSavedSessions()
	if len(savedSessions) != 1 {
		t.Fatalf("Expected one session to be saved (upon receiving the new token) but got %v", len(savedSessions))
	}
	firstSavedSession := savedSessions[0]
	if firstSavedSession.Token != token {
		t.Fatalf("Expected token '%v' to be saved but got '%v'", token, firstSavedSession.Token)
	}
}

func Test_SavedSession_Valid(t *testing.T) {
	token, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token: %v", err)
	}

	sessionCache := test_mocks.NewMockSessionCache(
		false,
		false,
		&session_cache.Session{Token: token},
	)

	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, token)

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, sessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err != nil {
		t.Fatalf("We expected a successful authorization, but an error was thrown: %v", err)
	}
	savedSessions := sessionCache.GetSavedSessions()
	if len(savedSessions) != 0 {
		t.Fatalf("Expected zero sessions to be saved (because a session was loaded from cache) but got %v", len(savedSessions))
	}
}

// This is the "user's on an airplane when their token expires" case
func Test_SavedSession_InGracePeriod_UnreachableAuth0(t *testing.T) {
	token, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		-1,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a test token: %v", err)
	}

	sessionCache := test_mocks.NewMockSessionCache(
		false,
		false,
		&session_cache.Session{Token: token},
	)

	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(true, token)

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, sessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err != nil {
		t.Fatalf("We expected a successful authorization due to being in the grace period (even though Auth0 is unreachable), but an error was thrown: %v", err)
	}
	savedSessions := sessionCache.GetSavedSessions()
	if len(savedSessions) != 0 {
		t.Fatalf("Expected zero sessions to be saved (because a session was loaded from cache) but got %v", len(savedSessions))
	}
}

func Test_SavedSession_InGracePeriod_ReachableAuth0(t *testing.T) {
	expiredToken, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		-1,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting an expired test token: %v", err)
	}

	sessionCache := test_mocks.NewMockSessionCache(
		false,
		false,
		&session_cache.Session{Token: expiredToken},
	)

	freshToken, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a fresh test token: %v", err)
	}

	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, freshToken)

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, sessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err != nil {
		t.Fatalf("We expected a successful authorization due to being in the grace period with reachable Auth0, but an error was thrown: %v", err)
	}
	savedSessions := sessionCache.GetSavedSessions()
	if len(savedSessions) != 1 {
		t.Fatalf("Expected 1 sessions to be saved (when we get the new token from Auth0) but got %v", len(savedSessions))
	}
	firstSavedSession := savedSessions[0]
	if firstSavedSession.Token != freshToken {
		t.Fatalf("Expected token '%v' but got token '%v'", freshToken, firstSavedSession.Token)
	}
}

func Test_SavedSession_BeyondGracePeriod_ReachableAuth0(t *testing.T) {
	expiredToken, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		-int(tokenExpirationGracePeriod.Seconds() + 1),
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting an expired test token: %v", err)
	}

	sessionCache := test_mocks.NewMockSessionCache(
		false,
		false,
		&session_cache.Session{Token: expiredToken},
	)

	freshToken, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		3600,
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting a fresh test token: %v", err)
	}

	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(false, freshToken)

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, sessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err != nil {
		t.Fatalf("We expected a successful authorization due to Auth0 being reachable even though the token is beyond the grace period, but an error was thrown: %v", err)
	}
	savedSessions := sessionCache.GetSavedSessions()
	if len(savedSessions) != 1 {
		t.Fatalf("Expected 1 sessions to be saved (when we get the new token from Auth0) but got %v", len(savedSessions))
	}
	firstSavedSession := savedSessions[0]
	if firstSavedSession.Token != freshToken {
		t.Fatalf("Expected token '%v' but got token '%v'", freshToken, firstSavedSession.Token)
	}
}

func Test_SavedSession_BeyondGracePeriod_UnreachableAuth0(t *testing.T) {
	expiredToken, err := createTestToken(
		testAuth0PrivateKey,
		auth0_constants.Audience,
		auth0_constants.Issuer,
		-int(tokenExpirationGracePeriod.Seconds() + 1),
		[]string{auth0_constants.ExecutionPermission},
	)
	if err != nil {
		t.Fatalf("An error occurred getting an expired test token: %v", err)
	}

	sessionCache := test_mocks.NewMockSessionCache(
		false,
		false,
		&session_cache.Session{Token: expiredToken},
	)

	deviceAuthorizer := test_mocks.NewMockDeviceAuthorizer(true, "")

	accessController := NewDeviceAuthAccessController(testAuth0PublicKeys, sessionCache, deviceAuthorizer)
	if err := accessController.Authorize(); err == nil {
		t.Fatalf("We expected authorization to be rejected due to having a token that's beyond the grace period with Auth0 unreachable, but authorization was allowed")
	}
}

// ================= HELPER FUNCTIONS ========================================================
func createTestToken(
		rsaPrivateKeyPem string,
		audience string,
		issuer string,
		expiresInSeconds int,
		permissions []string) (string, error) {
	now := time.Now()
	expiration := now.Add(time.Second * time.Duration(expiresInSeconds))
	claims := auth0_token_claims.Auth0TokenClaims{
		Audience:  audience,
		ExpiresAt: expiration.Unix(),
		IssuedAt:  now.Unix(),
		Issuer:    issuer,
		Scope:     "", // Unused
		Subject:   "", // Unused
		Permissions: permissions,
	}

	signingMethod := jwt.SigningMethodRS256
	token := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": signingMethod.Alg(),
			"kid": testAuth0KeyId,
		},
		Claims: claims,
		Method: signingMethod,
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(rsaPrivateKeyPem))
	if err != nil {
		return "", stacktrace.Propagate(err, "Error parsing private key from PEM string")
	}

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred signing the token")
	}
	return tokenStr, nil
}

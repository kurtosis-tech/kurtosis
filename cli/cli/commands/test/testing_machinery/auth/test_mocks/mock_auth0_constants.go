/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_mocks

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/auth0_token_claims"
	"github.com/palantir/stacktrace"
	"time"
)

const (
	TestAuth0KeyId = "test-key-id"
	TestAuth0PrivateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

)

var TestAuth0PublicKeys = map[string]string{
	TestAuth0KeyId: `-----BEGIN CERTIFICATE-----
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

// ================= HELPER FUNCTIONS ========================================================
func CreateTestToken(
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
			"kid": TestAuth0KeyId,
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

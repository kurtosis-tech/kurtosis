/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_constants

const (
	Audience = "https://api.kurtosistech.com/login"
	Issuer   = "https://dev-lswjao-7.us.auth0.com/"

	ExecutionScope = "execute:kurtosis-core"
)

/*
These come from https://dev-lswjao-7.us.auth0.com/.well-known/jwks.json
The reason they're hardcoded is because the user needs offline access, which means
 we can't always connect to Auth0 to pull the public keys. We could cache them for
 a time, but then we have to build public key caching-and-encryption machinery (and we'd
 need to encrypt because if the user can modify the public key, they can use their own
 private key and forge whatever tokens they please)

!!!! IMPORTANT NOTE !!!!!: We're taking this route because it's easier for now, but we'll
 need to change this code if we ever change our tenant's private key!! If we plan to be changing
 our keys often, we'll have to build the public key caching-and-encryption stuff after all.
 */
var RsaPublicKeyBase64 = map[string]string{
	"TXwelYLR6avnUg0WVU8Ar": "MIIDDTCCAfWgAwIBAgIJdiAMa56WI3XhMA0GCSqGSIb3DQEBCwUAMCQxIjAgBgNVBAMTGWRldi1sc3dqYW8tNy51cy5hdXRoMC5jb20wHhcNMjAwOTA0MTQyNzIxWhcNMzQwNTE0MTQyNzIxWjAkMSIwIAYDVQQDExlkZXYtbHN3amFvLTcudXMuYXV0aDAuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsbqjLh92qyPDmQzlmk7/ZvFwLYeD0vfysCXYkxLMNyOmzCSYgSOhAXjhP8D4TH2zhJv485efVIbnMNpU35PpzWN6K9Uc1Jfe7fUq5cb/VL+va1xzteKKjZwkPcoG9HCFltlBtx5TPFenXtvZLHPLt1hqHSamhzguK0kOX/6rChKK2eLyS+ynMLzjIBC8kUoFrMOlriEJHba2i55RGmQNlIGZwuY6EpHmylj48vEyX4YnBtRMQ1zsORp6MhDauCNtsv/2fptZ+yNStvxTN7O9Wucx2ur6dxByl1KE09jzxORZgQ37Uj0c4nenjHOZPWCom9bKb5JwjUAdwjC0rUus8QIDAQABo0IwQDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSHyc1Qw5MMl+Q0asMQiv7c6kNMuDAOBgNVHQ8BAf8EBAMCAoQwDQYJKoZIhvcNAQELBQADggEBAEtKXEryan4bhD1gKahb3dywElR2fJfZvudYXLPHmMBqr5CSef/awoC2j0zzih34vEEy1C5yXA4xERVf1HMGz4M/JE4ETIbOWG4f+KCLs8c74f0hyvQPijMzuYjqKDoDqmTk9zUk2N+S13iFL774r0ym46ZsGvDI36zfegrCBcfHBZrkggUOZlUaY3R7k8kHsyqVGA5ilebxhUGGtJiEKoFpbfCcPV9T4/3uaN45Ig/pGB86a46kp7SPEopjrdwXIBu0xx30Ln+u0i30uKiPNX0kc6wfxIsU8jWKEmbrYYozBUHUMAa4VmR8rz/xaxaScLzr/OCOkiuDU4Hi5qcPsus=",
	"9ys2AgsC9URKl4kPd6RVC": "MIIDDTCCAfWgAwIBAgIJLCtL1YEaC0KgMA0GCSqGSIb3DQEBCwUAMCQxIjAgBgNVBAMTGWRldi1sc3dqYW8tNy51cy5hdXRoMC5jb20wHhcNMjAwOTA0MTQyNzIxWhcNMzQwNTE0MTQyNzIxWjAkMSIwIAYDVQQDExlkZXYtbHN3amFvLTcudXMuYXV0aDAuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqS2BG8KMllVttp+f9/e/LaemGYSbf5e3DtVaDjiEv+W4kisBjyLslVgYgEKcLxewYeo6MZa1cXXba4mgJ27/jr43o+ovsDhyqPoOoDuU7JNwrqlh1pCEAM9JrQWzgYrIZ8OPxn4su+n3M5ZJNlUYrlL5oEF+l1JewS2VPxgp8x3cCNY4UqPs21tD7Egf2osyrh8fPmvuqZWraS8JD9W4Z6WMLWXg+4NrwumJoIKzLQPA1PXcQNROqrMUtF6YjkYS/oG3SFRn0tkfqZL2qvltsyoyFqytYr31SdJkiFiT1Nsg9+vfImQqKX+9PqlynjzBFAsH6M6NitNQVZBpNRTTgwIDAQABo0IwQDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTuk2WxbRdCypZzxPZne6eVzc4ouzAOBgNVHQ8BAf8EBAMCAoQwDQYJKoZIhvcNAQELBQADggEBAH5c1aHlHjqZUSB40B9gMsQBS0UQFBWpVDmQNM9HJPG/GF8PGihYXAUBsqLHlkg28UKj5x0QVEXELIPHIMAMa9SL6mAw7UFi9INE4EdGfjED1mfMU0S6cmMAQGNYgWislY4Mz1W82+J8f5u3viIscmjK4NFtm7B5+zmF5gXkT/NsV+++5MFtED51yCsaoTRFT3JUiAo2g+ZJlNP38nKfqTBo1Rp3QsOk6khdZmLQ8k8QkKrfTHbxyt/cokTk9wAOjrwufjKAwLoWH5iKUE0IZLAFxGY9W9ltIOZNDcqTNe4NV0NNI5nMv8AngioqvAiLooADfwisd+2WxVSTKokiuKo=",
}

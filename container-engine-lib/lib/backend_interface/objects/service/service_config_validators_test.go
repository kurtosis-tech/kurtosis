package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var testValidUserCustomLabels = map[string]string{
	"a":                  "a",
	"aa":                 "", // empty value is allowed
	"aaa":                "aaa",
	"a99a9":              "a99a9",
	"99aa9":              "99aa9",
	"a.7.3.5":            "a.7.3.5",
	"environment":        "RaNdoM", // allowed caps on values
	"system_environment": "random_random",
	"system.environment": "random.random",
	"system-environment": "random-random",
	"keys-with-63-chars-are-allowed-000000000000000000000000000000": "values-with-63-chars-are-allowed-000000000000000000000000000000",
}

var testInvalidUserCustomLabels = []map[string]string{
	// NOT VALID KEYS
	{
		"": "random", // empty key not allowed
	},
	{
		" ": "random", // whitespace key not allowed
	},
	{
		"SystemEnvironment": "random", // caps on key not allowed for both Docker and K8s
	},
	{
		"system/environment": "random", // slash on key not allowed for both Docker and K8s
	},
	{
		"_system_environment": "random", // not valid for K8s because it starts with a non-alphanumerical char
	},
	{
		"system_environment-": "random", // not valid for K8s because it ends with a non-alphanumerical char
	},
	{
		"more-than-63-chars-is-not-allowed-in-keys-0000000000000000000000": "random", // key exceeds max value
	},
	// NOT VALID VALUES
	{
		"system_environment": " ", // whitespace value not allowed
	},
	{
		"system_environment": "more-than-63-chars-is-not-allowed-in-values-00000000000000000000", // value exceeds max value
	},
	{
		"system_environment": "_random", // not valid for K8s because it starts with a non-alphanumerical char
	},
	{
		"system_environment": "random-", // not valid for K8s because it ends with a non-alphanumerical char
	},
}

func TestServiceConfigValidators(t *testing.T) {
	err := ValidateServiceConfigLabels(testValidUserCustomLabels)
	require.NoError(t, err)
	for _, notAllowedLabels := range testInvalidUserCustomLabels {
		err := ValidateServiceConfigLabels(notAllowedLabels)
		require.Error(t, err)
	}
}

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceNameValidation(t *testing.T) {
	validServiceNames := []string{
		"a",
		"abc",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef123", // 63 characters
		"a-b",
	}

	invalidServiceNames := []string{
		"1-geth-lighthouse", // 17 characters
		"-bc",
		"a--",
		"a_b",
		"a%b",
		"a:b",
		"a/b",
		"",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef1234", // 64 characters
	}

	for _, name := range validServiceNames {
		require.True(t, IsServiceNameValid(ServiceName(name)), "expected valid service name: %s", name)
	}

	for _, name := range invalidServiceNames {
		require.False(t, IsServiceNameValid(ServiceName(name)), "expected invalid service name: %s", name)
	}
}

func TestPortNameValidation(t *testing.T) {
	validPortNames := []string{
		"a",
		"abc",
		"abcdefabcdef123", // 15 characters
		"a-b",
	}

	invalidPortNames := []string{
		"1-dummy-port", // 12 characters
		"-bc",
		"a--",
		"a_b",
		"a%b",
		"a:b",
		"a/b",
		"",
		"abcdefabcdef1234", // 16 characters
	}

	for _, name := range validPortNames {
		require.True(t, IsPortNameValid(name), "expected valid port name: %s", name)
	}

	for _, name := range invalidPortNames {
		require.False(t, IsPortNameValid(name), "expected invalid port name: %s", name)
	}
}

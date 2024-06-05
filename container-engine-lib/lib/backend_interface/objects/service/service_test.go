package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidServiceName(t *testing.T) {
	require.True(t, IsServiceNameValid(ServiceName("a")))
	require.True(t, IsServiceNameValid(ServiceName("abc")))
	require.True(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef123"))) // 63 characters
}

func TestValidPortName(t *testing.T) {
	require.True(t, IsPortNameValid("a"))
	require.True(t, IsPortNameValid("abc"))
	require.True(t, IsPortNameValid("abcdefabcdef123")) // 15 characters
}

func TestInvalidServiceName(t *testing.T) {
	require.False(t, IsServiceNameValid(ServiceName("1-geth-lighthouse"))) // 17 characters
}

func TestInvalidPortName(t *testing.T) {
	require.False(t, IsPortNameValid("1-dummy-port")) // 12 characters
}

func TestServiceNameWithSpecialChars(t *testing.T) {
	require.True(t, IsServiceNameValid(ServiceName("a-b")))
	require.False(t, IsServiceNameValid(ServiceName("-bc")))
	require.False(t, IsServiceNameValid(ServiceName("a--")))
	require.False(t, IsServiceNameValid(ServiceName("a_b")))
	require.False(t, IsServiceNameValid(ServiceName("a%b")))
	require.False(t, IsServiceNameValid(ServiceName("a:b")))
	require.False(t, IsServiceNameValid(ServiceName("a/b")))
}

func TestPortNameWithSpecialChars(t *testing.T) {
	require.True(t, IsPortNameValid("a-b"))
	require.False(t, IsPortNameValid("-bc"))
	require.False(t, IsPortNameValid("a--"))
	require.False(t, IsPortNameValid("a_b"))
	require.False(t, IsPortNameValid("a%b"))
	require.False(t, IsPortNameValid("a:b"))
	require.False(t, IsPortNameValid("a/b"))
}

func TestServiceNameLength(t *testing.T) {
	require.False(t, IsServiceNameValid(ServiceName("")))
	require.True(t, IsServiceNameValid(ServiceName("a")))
	require.True(t, IsServiceNameValid(ServiceName("abc")))
	require.True(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef123")))
	require.False(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef1234")))
}

func TestPortNameLength(t *testing.T) {
	require.False(t, IsPortNameValid(""))
	require.True(t, IsPortNameValid("a"))
	require.True(t, IsPortNameValid("abc"))
	require.True(t, IsPortNameValid("abcdefabcdef123"))   // 15 characters
	require.False(t, IsPortNameValid("abcdefabcdef1234")) // 16 characters
}

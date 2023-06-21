package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidServiceName(t *testing.T) {
	require.True(t, IsServiceNameValid(ServiceName("a")))
	require.True(t, IsServiceNameValid(ServiceName("abc")))
	require.True(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef123")))
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

func TestServiceNameLength(t *testing.T) {
	require.False(t, IsServiceNameValid(ServiceName("")))
	require.True(t, IsServiceNameValid(ServiceName("a")))
	require.True(t, IsServiceNameValid(ServiceName("abc")))
	require.True(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef123")))
	require.False(t, IsServiceNameValid(ServiceName("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef1234")))
}

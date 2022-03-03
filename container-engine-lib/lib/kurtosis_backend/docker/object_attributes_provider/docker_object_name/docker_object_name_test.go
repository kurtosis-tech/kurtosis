package docker_object_name

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testNamesWithValidity = map[string]bool{
	"": false,
	" ": false,
	"a": true,
	"aaa": true,
	"FoOBaR": true,
	"aAa": true,
	"a99a9": true,
	"a.7.3.5": true,
	"2021-12-02_kurtosis-engine-server_82388239": true,
	"foo.bar": true,
	"foo-bar": true,
	"foo_bar": true,
	"foobar$": false,
	"foo,bar": false,
}

func TestEdgeCases(t *testing.T) {
	for value, shouldPass := range testNamesWithValidity {
		_, err := CreateNewDockerObjectName(value)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected object name string '%v' validity to be '%v' but was '%v'", shouldPass, didPass)
	}
}

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat("a", maxLength + 1)
	_, err := CreateNewDockerObjectName(invalidLabel)
	require.Error(t, err)
}

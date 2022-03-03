package docker_label_value

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testLabelValuesWithValidity = map[string]bool{
	"": true,
	" ": true,
	"a": true,
	"aaa": true,
	"aAa": true,
	"a99a9": true,
	"a.7.3.5": true,
	"my-port:8080/TCP,your-port:9090/TCP,his-port:9091/UDP,her-port:9091/TCP": true,
	"myPort.8080-TCP_yourPort.9090-TCP_hisPort.9091-UDP_herPort.9091-TCP": true,
}

func TestEdgeCases(t *testing.T) {
	for value, shouldPass := range testLabelValuesWithValidity {
		_, err := CreateNewDockerLabelValue(value)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected label value string '%v' validity to be '%v' but was '%v'", shouldPass, didPass)
	}
}

func TestTooLongValue(t *testing.T) {
	invalidLabel := strings.Repeat("a", maxLabelValueBytes + 1)
	_, err := CreateNewDockerLabelValue(invalidLabel)
	require.Error(t, err)
}

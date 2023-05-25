package kubernetes_annotation_value

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testAnnotationValuesWithValidity = map[string]bool{
	"":                       true,
	" ":                      true,
	"a":                      true,
	"aaa":                    true,
	"aAa":                    true,
	"a99a9":                  true,
	"a.7.3.5":                true,
	"port = / : barr::==foo": true,
	"my-port:8080/TCP,your-port:9090/TCP,his-port:9091/UDP,her-port:9091/TCP": true,
	"myPort.8080-TCP_yourPort.9090-TCP_hisPort.9091-UDP_herPort.9091-TCP":     true,
}

func TestEdgeCases(t *testing.T) {
	for value, shouldPass := range testAnnotationValuesWithValidity {
		_, err := CreateNewKubernetesAnnotationValue(value)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected annotation value string '%v' validity to be '%v' but was '%v'", value, shouldPass, didPass)
	}
}

func TestLongString(t *testing.T) {
	validLabel := strings.Repeat("a", 999)
	_, err := CreateNewKubernetesAnnotationValue(validLabel)
	require.NoError(t, err)
}

package image_org

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	for value, expected := range map[string]string{
		"":                 defaultOrg,
		"/":                defaultOrg,
		"ghcr.io/my-org":   "ghcr.io/my-org",
		"ghcr.io/my-org/":  "ghcr.io/my-org",
		"ghcr.io/my-org//": "ghcr.io/my-org",
	} {
		t.Setenv(EnvVarName, value)
		require.Equal(t, expected, Get(), "value %q", value)
	}
}

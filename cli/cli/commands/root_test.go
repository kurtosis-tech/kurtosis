package commands

import (
	"bytes"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	root := RootCmd
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, kurtosis_cli_version.KurtosisCLIVersion + "\n", buf.String())
}

// TODO More tests here, but have to figure out how to spin up a test engine that won't conflict with the real engine

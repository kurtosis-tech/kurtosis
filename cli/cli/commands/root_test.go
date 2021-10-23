package commands

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	cmd := RootCmd
	output := bytes.NewBufferString("")
	cmd.SetOut(output)
	cmd.SetArgs([]string{"version"})

	assert.Equal(t, output.String(), CliVersion)
}

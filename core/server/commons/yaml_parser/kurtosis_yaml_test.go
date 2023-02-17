package yaml_parser

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

var sampleCorrectYaml = []byte(`name: github.com/test/test`)
var sampleInCorrectYaml = []byte(`incorrect_name: github.com/test/test`)

func Test_parseKurtosisYamlInternal_Success(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleCorrectYaml, nil
	}

	path := "/root/kurtosis.yml"
	actual, err := parseKurtosisYamlInternal(path, mockRead)
	require.Nil(t, err)
	require.Equal(t, "github.com/test/test", actual.GetPackageName())
}

func Test_parseKurtosisYamlInternal_FailureWhileReading(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return nil, io.ErrClosedPipe
	}

	path := "/root/kurtosis.yml"
	_, err := parseKurtosisYamlInternal(path, mockRead)
	require.NotNil(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("Error occurred while reading the contents of %v", path))
}

func Test_parseKurtosisYamlInternal_IncorrectYaml(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleInCorrectYaml, nil
	}

	path := "/root/kurtosis.yml"
	_, err := parseKurtosisYamlInternal(path, mockRead)
	require.Nil(t, err)
	require.Equal(t, "", "")
}

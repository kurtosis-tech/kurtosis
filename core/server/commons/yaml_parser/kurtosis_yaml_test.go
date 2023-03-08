package yaml_parser

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

var (
	kurtosisYmlPath     = "/root/kurtosis.yml"
	sampleCorrectYaml   = []byte(`name: github.com/test-author/test-repo`)
	sampleInCorrectYaml = []byte(`incorrect_name_key: github.com/test/test`)
)

func Test_parseKurtosisYamlInternal_Success(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleCorrectYaml, nil
	}

	actual, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.Nil(t, err)
	require.Equal(t, "github.com/test-author/test-repo", actual.GetPackageName())
}

func Test_parseKurtosisYamlInternal_FailureWhileReading(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return nil, io.ErrClosedPipe
	}

	_, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.NotNil(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("Error occurred while reading the contents of '%v'", kurtosisYmlPath))
}

func Test_parseKurtosisYamlInternal_IncorrectYaml(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleInCorrectYaml, nil
	}

	actual, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.Nil(t, err)
	require.Equal(t, "", actual.GetPackageName())
}

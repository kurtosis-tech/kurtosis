package yaml_parser

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

var (
	kurtosisYmlPath                = "/root/kurtosis.yml"
	sampleCorrectYamlOnlyNameField = []byte(`
name: github.com/test-author/test-repo
`)
	sampleCompleteCorrectYaml = []byte(`
name: github.com/test-author/test-repo
description: |
  Some words to describe the package.
replace:
  github.com/kurtosis-tech/sample-dependency-package: github.com/kurtosis-tech/another-sample-dependency-package
  github.com/ethpandaops/ethereum-package: github.com/my-forked/ethereum-package
`)
	sampleInCorrectKeyYaml         = []byte(`incorrect_name_key: github.com/test/test`)
	sampleDuplicatedReplaceKeyYaml = []byte(`
name: github.com/test-author/test-repo
replace:
  github.com/kurtosis-tech/sample-dependency-package: github.com/kurtosis-tech/another-sample-dependency-package
  github.com/kurtosis-tech/sample-dependency-package: github.com/my-forked/ethereum-package
`)
)

func Test_parseKurtosisYamlInternal_Success(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleCompleteCorrectYaml, nil
	}

	actual, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.Nil(t, err)
	require.Equal(t, "github.com/test-author/test-repo", actual.GetPackageName())
}

func Test_parseKurtosisYamlInternal_OnlyNameSuccess(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleCorrectYamlOnlyNameField, nil
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

func Test_parseKurtosisYamlInternal_IncorrectKeyYaml(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleInCorrectKeyYaml, nil
	}

	_, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.Error(t, err)
	require.Contains(t, err.Error(), "incorrect_name_key not found")
}

func Test_parseKurtosisYamlInternal_DuplicatedReplaceKeyYaml(t *testing.T) {
	mockRead := func(filename string) ([]byte, error) {
		return sampleDuplicatedReplaceKeyYaml, nil
	}

	_, err := parseKurtosisYamlInternal(kurtosisYmlPath, mockRead)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key \"github.com/kurtosis-tech/sample-dependency-package\" already set in map")
}

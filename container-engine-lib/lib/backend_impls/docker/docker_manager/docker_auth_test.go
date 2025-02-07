package docker_manager

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeStaticConfig writes a static Docker config.json file to a temporary directory
func writeStaticConfig(t *testing.T, configContent string) string {
	tmpDir, err := os.MkdirTemp("", "docker-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// only write to file if content is not empty
	if configContent != "" {
		configPath := tmpDir + "/config.json"
		err = os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("Failed to write config.json: %v", err)
		}
	}

	// Set the DOCKER_CONFIG environment variable to the temp directory
	os.Setenv(ENV_DOCKER_CONFIG, tmpDir)
	return tmpDir
}

func TestGetAuthWithNoAuthSetReturnsNilAndNoError(t *testing.T) {
	// update docker config env var
	tmpDir := writeStaticConfig(t, "")
	defer os.RemoveAll(tmpDir)
	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.Nil(t, authConfig, "Auth config should be nil")
}

func TestGetAuthConfigForRepoPlain(t *testing.T) {
	expectedUser := "user"
	expectedPassword := "password"

	encodedAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedUser, expectedPassword)))

	cfg := fmt.Sprintf(`
	{
		"auths": {
			"https://index.docker.io/v1/": {
				"auth": "%s"
			}
		}
	}`, encodedAuth)

	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	// Test 1: Retrieve auth config for Docker Hub using docker.io domain
	authConfig, err := GetAuthFromDockerConfig("docker.io/my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.Equal(t, encodedAuth, authConfig.Auth, "Auth for Docker Hub should match")

	// Test 2: Retrieve auth config for Docker Hub using no domain
	authConfig, err = GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.Equal(t, encodedAuth, authConfig.Auth, "Auth for Docker Hub should match when using no host prefix")

	// Test 3: Retrieve auth config for Docker Hub using full domain and https:// prefix
	authConfig, err = GetAuthFromDockerConfig("https://registry-1.docker.io/my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.Equal(t, encodedAuth, authConfig.Auth, "Auth for Docker Hub should match when using no host prefix")

}

func TestGetAuthConfigForRepoOSX(t *testing.T) {
	t.Skip("Skipping test that requires macOS keychain")

	cfg := `{
		"auths": {
			"https://index.docker.io/v1/": {}
		},
		"credsStore": "osxkeychain"
	}`
	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
}

func TestGetAuthConfigForRepoUnix(t *testing.T) {
	t.Skip("Skipping test that requires unix `pass` password manager")

	cfg := `{
		"auths": {
			"https://index.docker.io/v1/": {}
		},
		"credsStore": "pass"
	}`
	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
}

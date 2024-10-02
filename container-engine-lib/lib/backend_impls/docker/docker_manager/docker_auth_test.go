package docker_manager

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/assert"
)

// WriteStaticConfig writes a static Docker config.json file to a temporary directory
func WriteStaticConfig(t *testing.T, configContent string) string {
	tmpDir, err := os.MkdirTemp("", "docker-config")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	configPath := tmpDir + "/config.json"
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config.json: %v", err)
	}

	// Set the DOCKER_CONFIG environment variable to the temp directory
	os.Setenv("DOCKER_CONFIG", tmpDir)
	return tmpDir
}

func TestGetAuthConfigForRepoPlain(t *testing.T) {
	expectedAuth := registry.AuthConfig{
		Username: "user",
		Password: "password",
	}

	encodedAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedAuth.Username, expectedAuth.Password)))

	cfg := fmt.Sprintf(`
	{
		"auths": {
			"https://index.docker.io/v1/": {
				"auth": "%s"
			}
		}
	}`, encodedAuth)

	tmpDir := WriteStaticConfig(t, cfg)
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
	tmpDir := WriteStaticConfig(t, cfg)
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
	tmpDir := WriteStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	authConfig, err := GetAuthFromDockerConfig("my-repo/my-image:latest")
	assert.NoError(t, err)
	assert.NotNil(t, authConfig, "Auth config should not be nil")
}

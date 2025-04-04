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
	expectedUserDockerHub := "dhuser"
	expectedPasswordDockerHub := "dhpassword"
	expectedUserGithub := "ghuser"
	expectedPasswordGithub := "ghpassword"

	encodedAuthDockerHub := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedUserDockerHub, expectedPasswordDockerHub)))
	encodedAuthGithub := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", expectedUserGithub, expectedPasswordGithub)))

	cfg := fmt.Sprintf(`
	{
		"auths": {
			"https://index.docker.io/v1/": {
				"auth": "%s"
			},
			"https://ghcr.io": {
				"auth": "%s"
			}
		}
	}`, encodedAuthDockerHub, encodedAuthGithub)

	tmpDir := writeStaticConfig(t, cfg)
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		repo         string
		expectedAuth string
	}{
		{
			repo:         "docker.io/my-repo/my-image:latest",
			expectedAuth: encodedAuthDockerHub,
		},
		{
			repo:         "my-repo/my-image:latest",
			expectedAuth: encodedAuthDockerHub,
		},
		{
			repo:         "https://registry-1.docker.io/my-repo/my-image:latest",
			expectedAuth: encodedAuthDockerHub,
		},
		{
			repo:         "https://index.docker.io/v1/",
			expectedAuth: encodedAuthDockerHub,
		},
		{
			repo:         "https://index.docker.io/v1",
			expectedAuth: encodedAuthDockerHub,
		},
		{
			repo:         "ghcr.io/my-repo/my-image:latest",
			expectedAuth: encodedAuthGithub,
		},
		{
			repo:         "ghcr.io",
			expectedAuth: encodedAuthGithub,
		},
		{
			repo:         "ghcr.io/",
			expectedAuth: encodedAuthGithub,
		},
	}

	for _, testCase := range testCases {
		authConfig, err := GetAuthFromDockerConfig(testCase.repo)
		assert.NoError(t, err)
		assert.NotNil(t, authConfig, "Auth config should not be nil")
		assert.Equal(t, testCase.expectedAuth, authConfig.Auth, "Auth for Docker Hub should match")
	}

	authConfig, err := GetAuthFromDockerConfig("something-else.local")
	assert.NoError(t, err)
	assert.Nil(t, authConfig, "Auth config should be nil")

	registries, err := GetAllRegistriesFromDockerConfig()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(registries))
	assert.Contains(t, registries, "https://index.docker.io/v1")
	assert.Contains(t, registries, "https://ghcr.io")
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

package docker_manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/registry"
	dockerregistry "github.com/docker/docker/registry"
)

// RegistryAuthConfig holds authentication configuration for a container registry
type RegistryAuthConfig struct {
	Auths       map[string]registry.AuthConfig `json:"auths"`
	CredHelpers map[string]string              `json:"credHelpers"`
	CredsStore  string                         `json:"credsStore"`
}

// loadDockerAuth loads the authentication configuration from the config.json file located in $DOCKER_CONFIG or ~/.docker
func loadDockerAuth() (RegistryAuthConfig, error) {
	configFilePath := os.Getenv("DOCKER_CONFIG")
	if configFilePath == "" {
		configFilePath = os.Getenv("HOME") + "/.docker/config.json"
	} else {
		configFilePath = configFilePath + "/config.json"
	}

	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return RegistryAuthConfig{}, fmt.Errorf("error reading Docker config file: %v", err)
	}

	var authConfig RegistryAuthConfig
	if err := json.Unmarshal(file, &authConfig); err != nil {
		return RegistryAuthConfig{}, fmt.Errorf("error unmarshalling Docker config file: %v", err)
	}

	return authConfig, nil
}

// getCredentialsFromStore fetches credentials from a Docker credential helper (credStore)
func getCredentialsFromStore(credHelper string, registryURL string) (*registry.AuthConfig, error) {
	// Prepare the helper command (docker-credential-<store>)
	credHelperCmd := "docker-credential-" + credHelper

	// Execute the credential helper to get credentials for the registry
	cmd := exec.Command(credHelperCmd, "get")
	cmd.Stdin = strings.NewReader(registryURL)

	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing credential helper %s: %v, %s", credHelperCmd, err, stderr.String())
	}

	// Parse the output (it should return JSON containing "Username", "Secret" and "ServerURL")
	creds := struct {
		Username  string `json:"Username"`
		Secret    string `json:"Secret"`
		ServerURL string `json:"ServerURL"`
	}{}

	if err := json.Unmarshal(out.Bytes(), &creds); err != nil {
		return nil, fmt.Errorf("error parsing credentials from store: %v", err)
	}

	return &registry.AuthConfig{
		Username:      creds.Username,
		Password:      creds.Secret,
		ServerAddress: creds.ServerURL,
		Auth:          "",
		Email:         "",
		IdentityToken: "",
		RegistryToken: "",
	}, nil
}

// GetAuthFromDockerConfig retrieves the auth configuration for a given repository
// by checking the Docker config.json file and Docker credential helpers.
// Returns nil if no credentials were found.
func GetAuthFromDockerConfig(repo string) (*registry.AuthConfig, error) {
	authConfig, err := loadDockerAuth()
	if err != nil {
		return nil, err
	}

	registryHost := dockerregistry.ConvertToHostname(repo)

	if !strings.Contains(registryHost, ".") || registryHost == "docker.io" || registryHost == "registry-1.docker.io" {
		registryHost = "https://index.docker.io/v1/"
	}

	// Check if the URL contains "://", meaning it already has a protocol
	if !strings.Contains(registryHost, "://") {
		registryHost = "https://" + registryHost
	}

	// 1. Check if there is a credHelper for this specific registry
	if credHelper, exists := authConfig.CredHelpers[registryHost]; exists {
		return getCredentialsFromStore(credHelper, registryHost)
	}

	// 2. Check if there is a default credStore for all registries
	if authConfig.CredsStore != "" {
		return getCredentialsFromStore(authConfig.CredsStore, registryHost)
	}

	// 3. Fallback to credentials in "auths" if no credStore is available
	if auth, exists := authConfig.Auths[registryHost]; exists {
		return &auth, nil
	}

	// Return no AuthConfig if no credentials were found
	return nil, nil
}

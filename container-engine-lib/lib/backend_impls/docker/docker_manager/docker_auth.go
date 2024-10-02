package docker_manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
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
func getCredentialsFromStore(credHelper string, registryURL string) (types.AuthConfig, error) {
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
		return registry.AuthConfig{}, fmt.Errorf("error executing credential helper %s: %v, %s", credHelperCmd, err, stderr.String())
	}

	// Parse the output (it should return JSON containing "Username" and "Secret")
	var creds registry.AuthConfig
	if err := json.Unmarshal(out.Bytes(), &creds); err != nil {
		return registry.AuthConfig{}, fmt.Errorf("error parsing credentials from store: %v", err)
	}

	return creds, nil
}

// getAuthFromDockerConfig retrieves the auth configuration for a given repository
func getAuthFromDockerConfig(repo string) (registry.AuthConfig, error) {
	authConfig, err := loadDockerAuth()
	if err != nil {
		return registry.AuthConfig{}, err
	}

	registryHost := dockerregistry.ConvertToHostname(repo)

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
		return auth, nil
	}

	// Return an empty AuthConfig if no credentials were found
	return registry.AuthConfig{}, nil
}

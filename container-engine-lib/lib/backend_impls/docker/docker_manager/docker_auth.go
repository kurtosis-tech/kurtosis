package docker_manager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/registry"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	ENV_DOCKER_CONFIG string = "DOCKER_CONFIG"
)

// RegistryAuthConfig holds authentication configuration for a container registry
type RegistryAuthConfig struct {
	Auths       map[string]registry.AuthConfig `json:"auths"`
	CredHelpers map[string]string              `json:"credHelpers"`
	CredsStore  string                         `json:"credsStore"`
}

// loadDockerAuth loads the authentication configuration from the config.json file located in $DOCKER_CONFIG or ~/.docker
func loadDockerAuth() (RegistryAuthConfig, error) {
	configFilePath := os.Getenv(ENV_DOCKER_CONFIG)
	if configFilePath == "" {
		configFilePath = os.Getenv("HOME") + "/.docker/config.json"
	} else {
		configFilePath = configFilePath + "/config.json"
	}

	file, err := os.ReadFile(configFilePath)
	if errors.Is(err, os.ErrNotExist) {
		// If the auth config doesn't exist, return an empty auth config
		logrus.Debugf("No docker config found at '%s'. Returning empty registry auth config.", configFilePath)
		return RegistryAuthConfig{}, nil
	} else if err != nil {
		return RegistryAuthConfig{}, stacktrace.Propagate(err, "error reading Docker config file at '%s'", configFilePath)
	}

	var authConfig RegistryAuthConfig
	if err := json.Unmarshal(file, &authConfig); err != nil {
		return RegistryAuthConfig{}, stacktrace.Propagate(err, "error unmarshalling Docker config file at '%s'", configFilePath)
	}

	return authConfig, nil
}

// getRegistriesFromCredsStore fetches all registries from a Docker credential helper (credStore)
func getRegistriesFromCredsStore(credHelper string) ([]string, error) {
	credHelperCmd := "docker-credential-" + credHelper

	cmd := exec.Command(credHelperCmd, "list")

	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, stacktrace.Propagate(err, "error executing credential helper '%s': %s", cmd.String(), stderr.String())
	}
	// Output will look like this: {"https://index.docker.io/v1/":"username"}
	var result map[string]string
	outStr := out.String()
	err := json.Unmarshal([]byte(outStr), &result)
	if err != nil {
		return nil, stacktrace.Propagate(err, "error unmarshaling credential helper list output '%s': %s", cmd.String(), outStr)
	}

	registries := []string{}
	for k := range result {
		registries = append(registries, k)
	}
	return registries, nil
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
		return nil, stacktrace.Propagate(err, "error executing credential helper '%s' for '%s': %s", cmd.String(), registryURL, stderr.String())
	}

	// Parse the output (it should return JSON containing "Username", "Secret" and "ServerURL")
	creds := struct {
		Username  string `json:"Username"`
		Secret    string `json:"Secret"`
		ServerURL string `json:"ServerURL"`
	}{}

	if err := json.Unmarshal(out.Bytes(), &creds); err != nil {
		return nil, stacktrace.Propagate(err, "error parsing credentials from store")
	}

	return &registry.AuthConfig{
		Username:      creds.Username,
		Password:      creds.Secret,
		ServerAddress: creds.ServerURL,
		Auth:          base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", creds.Username, creds.Secret))),
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

// GetAllRegistriesFromDockerConfig retrieves all registries from the Docker config.json file
func GetAllRegistriesFromDockerConfig() ([]string, error) {
	authConfig, err := loadDockerAuth()
	if err != nil {
		return nil, err
	}

	var registries []string
	for registry := range authConfig.Auths {
		registries = append(registries, registry)
	}

	for registry := range authConfig.CredHelpers {
		registries = append(registries, registry)
	}

	if authConfig.CredsStore != "" {
		r, err := getRegistriesFromCredsStore(authConfig.CredsStore)
		if err != nil {
			return nil, err
		}
		registries = append(registries, r...)
	}

	return registries, nil
}

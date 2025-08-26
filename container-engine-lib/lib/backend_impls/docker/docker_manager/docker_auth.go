package docker_manager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/registry"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
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

	logrus.Infof("Loading docker auth from config file: %s", configFilePath)

	file, err := os.ReadFile(configFilePath)
	if errors.Is(err, os.ErrNotExist) {
		// If the auth config doesn't exist, return an empty auth config
		logrus.Debugf("No docker config found at '%s'. Returning empty registry auth config.", configFilePath)
		return emptyRegistryAuthConfig(), nil
	} else if err != nil {
		return emptyRegistryAuthConfig(), stacktrace.Propagate(err, "error reading Docker config file at '%s'", configFilePath)
	}

	var authConfig RegistryAuthConfig
	if err := json.Unmarshal(file, &authConfig); err != nil {
		return emptyRegistryAuthConfig(), stacktrace.Propagate(err, "error unmarshalling Docker config file at '%s'", configFilePath)
	}

	// Remove trailing slashes from registry URLs
	for registry, auth := range authConfig.Auths {
		if strings.HasSuffix(registry, "/") {
			delete(authConfig.Auths, registry)
			registry = strings.TrimSuffix(registry, "/")
			authConfig.Auths[registry] = auth
		}
	}

	return authConfig, nil
}

func emptyRegistryAuthConfig() RegistryAuthConfig {
	return RegistryAuthConfig{
		Auths:       map[string]registry.AuthConfig{},
		CredHelpers: map[string]string{},
		CredsStore:  "",
	}
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

	// First try with the original URL
	auth, err := tryGetCredentialsWithURL(credHelperCmd, registryURL)
	if err == nil {
		return auth, nil
	}

	// If the URL doesn't end with a slash, try again with a trailing slash
	if !strings.HasSuffix(registryURL, "/") {
		auth, retryErr := tryGetCredentialsWithURL(credHelperCmd, registryURL+"/")
		if retryErr == nil {
			return auth, nil
		}
	}

	// If both attempts failed, return the original error
	return nil, err
}

// tryGetCredentialsWithURL attempts to get credentials for a specific URL
func tryGetCredentialsWithURL(credHelperCmd string, registryURL string) (*registry.AuthConfig, error) {
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
	}{
		Username:  "",
		Secret:    "",
		ServerURL: "",
	}

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

	// if repo string doesn't contain a repo prefix assume its an official docker library image
	if !strings.Contains(repo, "/") && !strings.Contains(repo, ".") {
		repo = "library/" + repo
	}

	// Remove tag from repo if it exists
	repo = strings.Split(repo, ":")[0]

	registryHost := dockerregistry.ConvertToHostname(repo)

	// Deal with the default Docker Hub registry.
	if !strings.Contains(registryHost, ".") ||
		registryHost == "docker.io" ||
		registryHost == "registry-1.docker.io" ||
		registryHost == "index.docker.io" {
		registryHost = "https://index.docker.io/v1"
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
		// Apply '/' to the end of the registry host if it doesn't have it
		if !strings.HasSuffix(registryHost, "/") {
			registryHost = registryHost + "/"
		}
		ac := registry.AuthConfig{
			ServerAddress: registryHost,
			Username:      auth.Username,
			Password:      auth.Password,
			Auth:          auth.Auth,
			Email:         auth.Email,
			IdentityToken: auth.IdentityToken,
			RegistryToken: auth.RegistryToken,
		}

		// If the username or password fields are set, set them in the AuthConfig (Overrides the decoded auth)
		if auth.Username != "" && auth.Password != "" {
			ac.Username = auth.Username
			ac.Password = auth.Password
		} else if auth.Auth != "" {
			// If the base64 encoded auth field is set, decode it and also set the Username and Password
			decodedAuth, err := base64.StdEncoding.DecodeString(auth.Auth)
			if err != nil {
				return nil, stacktrace.Propagate(err, "error decoding auth for registry '%s'", registryHost)
			}
			usernamePasswordSeparatorIndex := strings.IndexByte(string(decodedAuth), ':')
			if usernamePasswordSeparatorIndex != -1 {
				ac.Username = string(decodedAuth[:usernamePasswordSeparatorIndex])
				ac.Password = string(decodedAuth[usernamePasswordSeparatorIndex+1:])
			}
		} else {
			return nil, stacktrace.NewError("no username or password or auth found for registry '%s'", registryHost)
		}

		return &ac, nil
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

package update

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
)

func TestCreateUpdatedServiceConfigFromOverrides(t *testing.T) {
	testCases := []struct {
		name                  string
		currConfig            *services.ServiceConfig
		overrideConfig        *services.ServiceConfig
		expectedUpdatedConfig *services.ServiceConfig
	}{
		{
			name: "override image, entrypoint, cmd, env vars, ports, and files",
			currConfig: &services.ServiceConfig{
				Image:                       "old-image",
				PrivatePorts:                map[string]services.Port{"port1": {Number: 8080}},
				Files:                       map[string][]string{"/mnt/data": {"artifactA"}},
				Entrypoint:                  []string{"old-entry"},
				Cmd:                         []string{"old-cmd"},
				EnvVars:                     map[string]string{"FOO": "old"},
				PublicPorts:                 nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
			overrideConfig: &services.ServiceConfig{
				Image:                       "new-image",
				Entrypoint:                  []string{"new-entry"},
				Cmd:                         []string{"new-cmd"},
				PrivatePorts:                map[string]services.Port{"port2": {Number: 9090}},
				Files:                       map[string][]string{"/mnt/config": {"artifactB"}},
				EnvVars:                     map[string]string{"FOO": "new", "BAR": "added"},
				PublicPorts:                 nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
			expectedUpdatedConfig: &services.ServiceConfig{
				Image:                       "new-image",
				Entrypoint:                  []string{"new-entry"},
				Cmd:                         []string{"new-cmd"},
				PrivatePorts:                map[string]services.Port{"port1": {Number: 8080}, "port2": {Number: 9090}},
				Files:                       map[string][]string{"/mnt/data": {"artifactA"}, "/mnt/config": {"artifactB"}},
				EnvVars:                     map[string]string{"FOO": "new", "BAR": "added"},
				PublicPorts:                 nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
		},
		{
			name: "no overrides applied",
			currConfig: &services.ServiceConfig{
				Image:                       "base-image",
				Entrypoint:                  []string{"entry"},
				Cmd:                         []string{"cmd"},
				PrivatePorts:                map[string]services.Port{"http": {Number: 80}},
				Files:                       map[string][]string{"/data": {"foo"}},
				EnvVars:                     map[string]string{"KEY": "VAL"},
				PublicPorts:                 nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
			overrideConfig: &services.ServiceConfig{},
			expectedUpdatedConfig: &services.ServiceConfig{
				Image:                       "base-image",
				Entrypoint:                  []string{"entry"},
				Cmd:                         []string{"cmd"},
				PrivatePorts:                map[string]services.Port{"http": {Number: 80}},
				Files:                       map[string][]string{"/data": {"foo"}},
				EnvVars:                     map[string]string{"KEY": "VAL"},
				PublicPorts:                 nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
		},
		{
			name: "override overwrites duplicate env var and port",
			currConfig: &services.ServiceConfig{
				Image:                       "original",
				PrivatePorts:                map[string]services.Port{"p": {Number: 1000}},
				EnvVars:                     map[string]string{"K1": "V1"},
				Files:                       map[string][]string{},
				PublicPorts:                 nil,
				Entrypoint:                  nil,
				Cmd:                         nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
			overrideConfig: &services.ServiceConfig{
				Image:                       "",
				PrivatePorts:                map[string]services.Port{"p": {Number: 2000}},
				EnvVars:                     map[string]string{"K1": "override"},
				Files:                       map[string][]string{},
				PublicPorts:                 nil,
				Entrypoint:                  nil,
				Cmd:                         nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
			expectedUpdatedConfig: &services.ServiceConfig{
				Image:                       "original",
				PrivatePorts:                map[string]services.Port{"p": {Number: 2000}},
				EnvVars:                     map[string]string{"K1": "override"},
				Files:                       map[string][]string{},
				PublicPorts:                 nil,
				Entrypoint:                  nil,
				Cmd:                         nil,
				PrivateIPAddressPlaceholder: "",
				MaxMillicpus:                0,
				MinMillicpus:                0,
				MaxMemory:                   0,
				MinMemory:                   0,
				User:                        nil,
				Tolerations:                 nil,
				Labels:                      nil,
				NodeSelectors:               nil,
				TiniEnabled:                 nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			updated := createUpdatedServiceConfigFromOverrides(tc.overrideConfig, tc.currConfig)
			require.Equal(t, tc.expectedUpdatedConfig.Image, updated.Image)
			require.Equal(t, tc.expectedUpdatedConfig.Entrypoint, updated.Entrypoint)
			require.Equal(t, tc.expectedUpdatedConfig.Cmd, updated.Cmd)
			require.Equal(t, tc.expectedUpdatedConfig.EnvVars, updated.EnvVars)
			require.Equal(t, tc.expectedUpdatedConfig.PrivatePorts, updated.PrivatePorts)
			require.Equal(t, tc.expectedUpdatedConfig.Files, updated.Files)
		})
	}
}

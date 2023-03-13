package backend_interface

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"path"
	"plugin"
)

const (
	KubernetesPluginName   = "kubernetes"
	pluginEntrypointSymbol = "Plugin"
)

type BackendPlugin interface {
	GetCLIBackend(ctx context.Context) (KurtosisBackend, error)
	GetEngineServerBackend(ctx context.Context) (KurtosisBackend, error)
	GetApiContainerBackend(ctx context.Context) (KurtosisBackend, error)
}

func OpenBackendPlugin(pluginPath string) (BackendPlugin, error) {
	if _, err := os.Stat(pluginPath); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"This backend requires a plugin in path '%s' that couldn't be found",
			pluginPath,
		)
	}
	pluginFile, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"This backend requires a plugin in path '%s' that couldn't be open",
			pluginPath,
		)
	}
	symbol, err := pluginFile.Lookup(pluginEntrypointSymbol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when looking up plugin entry point '%s'",
			pluginEntrypointSymbol,
		)
	}
	plugin, ok := symbol.(BackendPlugin)
	if !ok {
		return nil, stacktrace.NewError(
			"An error occurred casting a plugin backend loaded from plugin on path '%s'",
			pluginPath,
		)
	}
	return plugin, nil
}

func GetPluginPathForCLI(pluginName string) string {
	return path.Join(os.Getenv("HOME"), ".kurtosis", "plugins", fmt.Sprintf("%s.so", pluginName))

}

func GetPluginPathForEngine(pluginName string) string {
	return path.Join(GetPluginDirForEngine(), fmt.Sprintf("%s-engine-apic.so", pluginName))
}

func GetPluginPathForApiContainer(pluginName string) string {
	return path.Join(GetPluginDirForApiContainer(), fmt.Sprintf("%s-engine-apic.so", pluginName))
}

func GetPluginDirForApiContainer() string {
	return path.Join("/", "tmp", "plugins")
}

func GetPluginDirForEngine() string {
	return path.Join("/", "tmp", "plugins")
}

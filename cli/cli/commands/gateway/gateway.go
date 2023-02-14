package gateway

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"plugin"
)

const (
	emptyConfigMasterUrl       = ""
	runEngineGatewaySymbolName = "RunEngineGateway"
)

// GatewayCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var GatewayCmd = &cobra.Command{
	Use:   command_str_consts.GatewayCmdStr,
	Short: "Starts a local gateway to a Kurtosis cluster running in Kubernetes",
	RunE:  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get Kurtosis cluster configuration, instead a non-nil error was returned")
	}

	kurtosisBackend, err := clusterConfig.GetKurtosisBackend(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get a Kurtosis backend connected to the cluster, instead a non-nil error was returned")
	}

	// TODO Store kube config path in configuration and read from there
	kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube", "config")

	kubernetesConfig, err := clientcmd.BuildConfigFromFlags(emptyConfigMasterUrl, kubeConfigPath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kubernetes configuration from flags in file '%v'", kubeConfigPath)
	}

	pluginPath := backend_interface.GetPluginPathForCLI(backend_interface.KubernetesPluginName)
	pluginFile, err := plugin.Open(pluginPath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening Kubernetes plugin on path '%s'", pluginPath)
	}
	runEngineGatewaySymbol, err := pluginFile.Lookup(runEngineGatewaySymbolName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred looking up symbol '%s'  from Kubernetes plugin (path '%s')", runEngineGatewaySymbol, pluginPath)
	}
	runEngineGateway, ok := runEngineGatewaySymbol.(func(ctx context.Context, kubernetesConfig *rest.Config, kurtosisBackend backend_interface.KurtosisBackend) error)
	if !ok {
		return stacktrace.NewError("An error occurred when parsing gateway function from plugin")
	}
	return runEngineGateway(cmd.Context(), kubernetesConfig, kurtosisBackend)
}

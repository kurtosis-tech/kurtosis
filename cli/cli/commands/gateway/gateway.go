package gateway

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/run/engine_gateway"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
)

const (
	emptyConfigMasterUrl = ""
)

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

	connectionProvider, err := connection.NewGatewayConnectionProvider(cmd.Context(), kubernetesConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to instantiate a gateway connection provider, instead a non-nil error was returned")
	}

	if err := engine_gateway.RunEngineGatewayUntilInterrupted(kurtosisBackend, connectionProvider); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the engine gateway server.")
	}

	return nil
}

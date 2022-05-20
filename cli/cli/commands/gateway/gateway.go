package gateway

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/run/engine_gateway"
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

	runningEngines, err := kurtosisBackend.GetEngines(cmd.Context(), getRunningEnginesFilter())
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get engines that are running in Kurtosis, instead a non-nil err was returned")
	}
	if len(runningEngines) != 1 {
		return stacktrace.NewError("Expected to find exactly 1 running engine in Kurtosis, instead found '%v'", len(runningEngines))
	}
	runningEngine := getFirstEngineFromMap(runningEngines)
	// If the engine is running in kubernetes, there's no portspec for the public port

	if err := engine_gateway.RunEngineGatewayUntilInterrupted(runningEngine, connectionProvider); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the engine gateway server.")
	}

	return nil
}

func getRunningEnginesFilter() *engine.EngineFilters {
	return &engine.EngineFilters{
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Running: true,
		},
	}
}

// getFirstEngineFromMap returns the first value iterated by the `range` statement on a map
// returns nil if the map is empty
func getFirstEngineFromMap(engineMap map[string]*engine.Engine) *engine.Engine {
	firstEngineInMap := (*engine.Engine)(nil)
	for _, engineInMap := range engineMap {
		firstEngineInMap = engineInMap
		break
	}
	return firstEngineInMap
}

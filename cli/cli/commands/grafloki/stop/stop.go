package stop

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/grafloki"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

var GraflokiStopCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.GraflokiStopCmdStr,
	ShortDescription: "Stops a grafana/loki instance.",
	LongDescription:  "Stop a grafana/loki instance if one already exists.",
	RunFunc:          run,
}

func run(
	ctx context.Context,
	_ *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
	dockerManager, err := docker_manager.CreateDockerManager(grafloki.EmptyDockerClientOpts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Docker manager.")
	}
	if err := dockerManager.RemoveContainer(ctx, grafloki.GrafanaContainerName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing Grafana container '%v'", grafloki.GrafanaContainerName)
	}
	if err := dockerManager.RemoveContainer(ctx, grafloki.LokiContainerName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing Loki container '%v'", grafloki.GrafanaContainerName)
	}
	fmt.Println("Successfully stopped Grafana and Loki containers.")
	return nil
}

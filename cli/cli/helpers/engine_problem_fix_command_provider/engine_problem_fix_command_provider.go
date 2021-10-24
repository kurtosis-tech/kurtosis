package engine_problem_fix_command_provider

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/start"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/stop"
	"os"
	"path"
)

// Command to run if no engine is running but one should be
func GetNoEngineRunningCmd() string {
	return fmt.Sprintf(
		"%v %v %v",
		getBinaryFilename(),
		engine.CommandStr,
		start.CommandStr,
	)
}

// Command to run if an engine container is running but the server inside isn't responding
func GetEngineRunningButServerNotRespondingCmd() string {
	binaryFilename := getBinaryFilename()
	return fmt.Sprintf(
		"%v %v %v && %v %v 5v",
		binaryFilename,
		engine.CommandStr,
		stop.CommandStr,
		binaryFilename,
		engine.CommandStr,
		start.CommandStr,
	)

}

func getBinaryFilename() string {
	return path.Base(os.Args[0])
}

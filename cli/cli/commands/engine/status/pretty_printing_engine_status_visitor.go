package status

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/sirupsen/logrus"
)

const (
	engineVersionInfoLabel = "Version"
)

// Pretty printer of engine status that will compile-break any time a new engine status is added
type prettyPrintingEngineStatusVisitor struct {
	// Will only be filled in if the engine is running
	maybeApiVersion string
}

func newPrettyPrintingEngineStatusVisitor(maybeApiVersion string) *prettyPrintingEngineStatusVisitor {
	return &prettyPrintingEngineStatusVisitor{maybeApiVersion: maybeApiVersion}
}

func (p *prettyPrintingEngineStatusVisitor) VisitStopped() error {
	fmt.Fprintln(logrus.StandardLogger().Out, "No Kurtosis engine is running")
	return nil
}

func (p *prettyPrintingEngineStatusVisitor) VisitContainerRunningButServerNotResponding() error {
	fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine container is running, but the server inside couldn't be reached")
	return nil
}

func (p *prettyPrintingEngineStatusVisitor) VisitRunning() error {
	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(engineVersionInfoLabel, p.maybeApiVersion)

	fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine is running with the following info:")
	keyValuePrinter.Print()
	return nil
}


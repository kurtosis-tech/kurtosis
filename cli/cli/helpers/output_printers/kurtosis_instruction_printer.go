package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"strings"
)

const (
	bazelBuildDefaultFilename = ""
)

func FormatInstruction(instruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction) string {
	canonicalizedInstructionWithComment := fmt.Sprintf(
		"# from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)

	parsedInstruction, _ := build.ParseDefault(bazelBuildDefaultFilename, []byte(canonicalizedInstructionWithComment))

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}

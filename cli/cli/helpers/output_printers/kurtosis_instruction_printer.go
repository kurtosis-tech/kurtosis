package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/sirupsen/logrus"
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

	parsedInstruction, err := build.ParseDefault(bazelBuildDefaultFilename, []byte(canonicalizedInstructionWithComment))
	if err != nil {
		logrus.Warnf("Unable to format instruction. Will print it with no indentation. Problematic instruction was: \n%v", canonicalizedInstructionWithComment)
		return canonicalizedInstructionWithComment
	}

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}

package startosis_warning

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type WarningStruct struct {
	stream grpc.ServerStream
	isOpen bool
}

var warn WarningStruct

func Setup(stream grpc.ServerStream) {
	if !warn.isOpen {
		warn = WarningStruct{stream: stream}
		warn.isOpen = true
	}
}

// Close this method is called once the starlark execution is completed
func Close() {
	warn.isOpen = false
	warn.stream = nil
}

func Printf(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	if warn.stream != nil {
		if err := warn.stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(formattedMessage)); err != nil {
			logrus.Errorf("Error Occurred while streaming warning message to the cli %q .Error: %+v", formattedMessage, err)
		}
	}
}

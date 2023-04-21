package startosis_warning

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sync"
)

type WarningStruct struct {
	stream grpc.ServerStream
}

var once *sync.Once
var warn WarningStruct
var lock = &sync.Mutex{}

//Setup initialize the warning struct with the current available grpc stream
func Setup(stream grpc.ServerStream) {
	if once == nil {
		once = &sync.Once{}
	}

	once.Do(func() {
		warn = WarningStruct{stream: stream}
	})
}

// Close this method is called once the starlark execution is completed
// This method is called during clean up and ensures that next time
// when set up is called - warn struct will be initialized with current available stream
func Close() {
	once = nil
}

//TODO: add an abstraction so that this method actually prints to cli only once!
func PrintOncef(message string, args ...interface{}) {
	lock.Lock()
	defer lock.Unlock()

	formattedMessage := fmt.Sprintf(message, args...)
	if warn.stream != nil {
		if err := warn.stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(formattedMessage)); err != nil {
			logrus.Errorf("Error Occurred while streaming warning message to the cli %q .Error: %+v", formattedMessage, err)
		}
	}
}

package common

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
)

func ForwardKurtosisExecutionStream[T any](streamToReadFrom grpc.ClientStream, streamToWriteTo grpc.ServerStream) error {
	for {
		starlarkRunResponseLine := new(T)
		// RecvMsg blocks until either a message is received or an error is thrown
		readErr := streamToReadFrom.RecvMsg(starlarkRunResponseLine)
		if readErr == io.EOF {
			logrus.Debug("Finished reading from the Kurtosis response line stream.")
			return nil
		}
		if readErr != nil {
			return stacktrace.Propagate(readErr, "Error reading Kurtosis execution lines from Kurtosis core stream")
		}

		if writeErr := streamToWriteTo.SendMsg(starlarkRunResponseLine); writeErr != nil {
			return stacktrace.Propagate(writeErr, "Received a Kurtosis execution line but failed forwarding it back to the user")
		}
	}
}

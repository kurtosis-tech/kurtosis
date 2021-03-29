/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package output

import (
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type streamerState string

const (
	// How long for the streamer to wait between copying the latest test output to system output
	timeBetweenStreamerCopies = 500 * time.Millisecond

	// If we ask the streamer to stop and it hasn't after this time, throw an error
	streamerStopTimeout = 5 * time.Second

	notStarted streamerState = "NOT_STARTED"
	streaming streamerState = "STREAMING"
	terminated streamerState = "TERMINATED"
	failedToStop streamerState = "FAILED_TO_STOP"
)

// Single-use, non-thread-safe streamer that will read data and pump it to the given output log
type LogStreamer struct {
	state streamerState

	// A channel to tell the streaming thread to stop
	// Will be set to non-nil when streaming starts
	streamThreadShutdownChan chan bool

	// A channel to indicate that the streaming thread has stopped
	// Will be set to non-nil when streaming starts
	streamThreadStoppedChan chan bool

	// Output logger to stream to
	outputLogger *logrus.Logger

	// Hook that will be called after the streaming thread is shutdown
	threadShutdownHook func()
}

func NewLogStreamer(outputLogger *logrus.Logger) *LogStreamer {
	return &LogStreamer{
		state:                    notStarted,
		streamThreadShutdownChan: nil,
		streamThreadStoppedChan:  nil,
		outputLogger:             outputLogger,
	}
}

func (streamer *LogStreamer) StartStreamingFromFilepath(inputFilepath string) error {
	input, err := os.Open(inputFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening input filepath '%v' for reading", inputFilepath)
	}

	threadShutdownHook := func() {
		input.Close()
	}

	if err := streamer.startStreamingThread(input, false, threadShutdownHook); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the streaming thread from filepath '%v'", inputFilepath)
	}
	return nil
}

func (streamer *LogStreamer) StartStreamingFromDockerLogs(input io.Reader) error {
	threadShutdownHook := func() {}
	if err := streamer.startStreamingThread(input, true, threadShutdownHook); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the streaming thread from the given reader")
	}
	return nil
}

func (streamer *LogStreamer) StopStreaming() error {
	// Stop is idempotent
	streamer.outputLogger.Tracef("[STREAMER] Received 'stop streaming' command while in state '%v'...", streamer.state)
	if streamer.state == terminated || streamer.state == failedToStop {
		streamer.outputLogger.Tracef("[STREAMER] Short-circuiting stop; streamer state is already '%v' state", streamer.state)
		return nil
	}
	if streamer.state != streaming {
		return stacktrace.NewError("Cannot stop streamer; streamer is not in 'streaming' state")
	}

	streamer.outputLogger.Trace("[STREAMER] Sending signal to stop streaming thread...")
	streamer.streamThreadShutdownChan <- true
	streamer.outputLogger.Trace("[STREAMER] Successfully sent signal to stop streaming thread")

	streamer.outputLogger.Tracef("[STREAMER] Waiting until thread reports stopped, or %v timeout is hit...", streamerStopTimeout)
	select {
	case <- streamer.streamThreadStoppedChan:
		streamer.outputLogger.Tracef("[STREAMER] Thread reported stop")
		streamer.state = terminated
		return nil
	case <- time.After(streamerStopTimeout):
		streamer.outputLogger.Tracef("[STREAMER] %v timeout was hit", streamerStopTimeout)
		streamer.state = failedToStop
		return stacktrace.NewError("We asked the streamer to stop but it still hadn't after %v", streamerStopTimeout)
	}
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (streamer *LogStreamer) startStreamingThread(input io.Reader, useDockerDemultiplexing bool, threadShutdownHook func()) error {
	if streamer.state != notStarted {
		return stacktrace.NewError("Cannot start streaming with this log streamer; streamer is not in the '%v' state", notStarted)
	}

	streamThreadShutdownChan := make(chan bool)
	streamThreadStoppedChan := make(chan bool)

	streamer.streamThreadShutdownChan = streamThreadShutdownChan
	streamer.streamThreadStoppedChan = streamThreadStoppedChan

	go func() {
		defer threadShutdownHook()

		keepGoing := true
		for keepGoing {
			select {
			case <- streamer.streamThreadShutdownChan:
				keepGoing = false
			case <- time.After(timeBetweenStreamerCopies):
				if err := copyToOutput(input, streamer.outputLogger.Out, useDockerDemultiplexing); err != nil {
					streamer.outputLogger.Errorf("An error occurred copying the output from the test logs: %v", err)
				}
			}
		}
		if err := copyToOutput(input, streamer.outputLogger.Out, useDockerDemultiplexing); err != nil {
			streamer.outputLogger.Error("An error occurred copying the final output from the test logs")
		}
		streamer.streamThreadStoppedChan <- true
	}()
	streamer.state = streaming
	return nil
}

func copyToOutput(input io.Reader, output io.Writer, useDockerDemultiplexing bool) error {
	var result error
	if useDockerDemultiplexing {
		_, result = stdcopy.StdCopy(output, output, input)
	} else {
		_, result = io.Copy(output, input)
	}
	return result
}


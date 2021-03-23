/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package output

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type streamerState int

const (
	// How long for the streamer to wait between copying the latest test output to system output
	timeBetweenStreamerCopies = 1 * time.Second

	// If we ask the streamer to stop and it hasn't after this time, throw an error
	streamerStopTimeout = 5 * time.Second

	notStarted streamerState = iota
	streaming
	terminated
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

	if err := streamer.startStreamingThread(input, threadShutdownHook); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the streaming thread from filepath '%v'", inputFilepath)
	}
	return nil
}

func (streamer *LogStreamer) StartStreamingFromReader(input io.Reader) error {
	threadShutdownHook := func() {}
	if err := streamer.startStreamingThread(input, threadShutdownHook); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the streaming thread from the given reader")
	}
	return nil
}

func (streamer *LogStreamer) StopStreaming() error {
	// Stop is idempotent
	if streamer.state == terminated {
		return nil
	}
	if streamer.state != streaming {
		return stacktrace.NewError("Cannot stop streamer; streamer is not in 'streaming' state")
	}

	streamer.streamThreadShutdownChan <- true

	select {
	case <- streamer.streamThreadStoppedChan:
		return nil
	case <- time.After(streamerStopTimeout):
		return stacktrace.NewError("We asked the streamer to stop but it still hadn't after %v", streamerStopTimeout)
	}
	streamer.state = terminated
	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (streamer *LogStreamer) startStreamingThread(input io.Reader, threadShutdownHook func()) error {
	if streamer.state != notStarted {
		return stacktrace.NewError("Cannot start streaming with this log streamer; streamer is not in the 'not started' state")
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
				if _, err := io.Copy(streamer.outputLogger.Out, input); err != nil {
					streamer.outputLogger.Error("An error occurred copying the output from the test logs")
				}
			}
		}
		if _, err := io.Copy(streamer.outputLogger.Out, input); err != nil {
			streamer.outputLogger.Error("An error occurred copying the final output from the test logs")
		}
		streamer.streamThreadStoppedChan <- true
	}()
	streamer.state = streaming
	return nil
}


/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

const (
	// How long for the streamer to wait between copying the latest test output to system output
	timeBetweenStreamerCopies = 1 * time.Second

	// If we ask the streamer to stop and it hasn't after this time, throw an error
	streamerStopTimeout = 5 * time.Second
)

// Single-use streamer that will read data from the given input log filepath, and pump it to the given output log
type logStreamer struct {
	// A channel to tell the streaming thread to stop
	shutdownChan chan bool

	// A channel to indicate that the streamer has stopped
	streamerStoppedChan chan bool

	inputLogFilepath string

	outputLogger *logrus.Logger
}

func newLogStreamer(inputLogFilepath string, outputLogger *logrus.Logger) *logStreamer {
	return &logStreamer{
		shutdownChan:        make(chan bool),
		streamerStoppedChan: make(chan bool),
		inputLogFilepath: inputLogFilepath,
		outputLogger:   outputLogger,
	}
}

func (streamer *logStreamer) startStreaming() error {
	inputLogReader, err := os.Open(streamer.inputLogFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening input log '%v' for reading", streamer.inputLogFilepath)
	}

	// NOTE: We intentionally alter the internal state of this object ONLY AFTER the last possible error, to
	// make this method idempotent in the event of an error
	go func() {
		defer inputLogReader.Close()

		keepGoing := true
		for keepGoing {
			select {
			case <- streamer.shutdownChan:
				keepGoing = false
			case <- time.After(timeBetweenStreamerCopies):
				if _, err := io.Copy(streamer.outputLogger.Out, inputLogReader); err != nil {
					streamer.outputLogger.Error("An error occurred copying the output from the test logs")
				}
			}
		}
		if _, err := io.Copy(streamer.outputLogger.Out, inputLogReader); err != nil {
			streamer.outputLogger.Error("An error occurred copying the final output from the test logs")
		}
		streamer.streamerStoppedChan <- true
	}()
	return nil
}

func (streamer *logStreamer) stopStreaming() error {
	streamer.shutdownChan <- true

	select {
	case <- streamer.streamerStoppedChan:
		// Nothing to do
	case <- time.After(streamerStopTimeout):
		return stacktrace.NewError("We asked the streamer to stop but it still hadn't after %v", streamerStopTimeout)
	}

	return nil
}


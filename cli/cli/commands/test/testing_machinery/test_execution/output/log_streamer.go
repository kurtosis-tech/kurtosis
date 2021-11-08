/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package output

import (
	"context"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type streamerState string

type streamerInputSource string

const (
	// How long the streamer will pause between each cycle of copying input -> output
	timeBetweenStreamerCopies = 500 * time.Millisecond

	// If we ask the streamer to stop and it hasn't after this time, throw an error
	streamerStopTimeout = 5 * time.Second

	notStarted   streamerState = "NOT_STARTED"
	streaming    streamerState = "STREAMING"
	terminated   streamerState = "TERMINATED"
	failedToStop streamerState = "FAILED_TO_STOP"

	inputFilePath streamerInputSource = "FILE_POINTER"
	inputDocker   streamerInputSource = "DOCKER_LOG_STREAM"

	shouldFollowTestsuiteLogs = true
)

// Single-use, non-thread-safe streamer that will read data and pump it to the given output log
type LogStreamer struct {
	// The label to give loglines originating from inside this logger
	loglineLabel string

	state streamerState

	inputState streamerInputSource

	// A channel to tell the streaming thread to stop
	// Will be set to non-nil when streaming starts
	filePathStreamThreadShutdownChan chan bool

	// A channel to indicate that the streaming thread has stopped
	// Will be set to non-nil when streaming starts
	streamThreadStoppedChan chan bool

	// Output logger to stream to
	outputLogger *logrus.Logger

	// Hook that will be called after the streaming thread is shutdown
	threadShutdownHook func()

	//Pointer to ReadCloser that has been opened by Docker input on this struct
	//Will be nil if non-Docker input however
	inputReadCloser io.ReadCloser
}

func NewLogStreamer(loglineLabel string, outputLogger *logrus.Logger) *LogStreamer {
	return &LogStreamer{
		loglineLabel:                     loglineLabel,
		state:                            notStarted,
		inputState:                       inputFilePath,
		filePathStreamThreadShutdownChan: nil,
		streamThreadStoppedChan:          nil,
		outputLogger:                     outputLogger,
		inputReadCloser:                  nil,
	}
}

func (streamer *LogStreamer) StartStreamingFromFilepath(inputFilepath string) error {

	streamer.inputState = inputFilePath

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

func (streamer *LogStreamer) StartStreamingFromDockerLogs(testSetupExecutionCtx context.Context,
	dockerManager *docker_manager.DockerManager, testsuiteContainerId string) error {

	streamer.inputState = inputDocker

	functionExitedSuccessfully := false
	input, err := dockerManager.GetContainerLogs(testSetupExecutionCtx, testsuiteContainerId, shouldFollowTestsuiteLogs)


	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the testsuite container logs for streaming.")
	} else {

		defer func() {
			if !functionExitedSuccessfully {
				input.Close()
			}
		}()

		streamer.inputReadCloser = input

		threadShutdownHook := func() {
			input.Close()
		}
		if err := streamer.startStreamingThread(input, threadShutdownHook); err != nil {
			return stacktrace.Propagate(err, "An error occurred starting the streaming thread from the given reader")
		}


	}

	functionExitedSuccessfully = true

	return nil
}

func (streamer *LogStreamer) StopStreaming() error {
	// Stop is idempotent
	streamer.outputLogger.Tracef("%vReceived 'stop streaming' command while in state '%v'...", streamer.getLoglinePrefix(), streamer.state)
	if streamer.state == terminated || streamer.state == failedToStop {
		streamer.outputLogger.Tracef("%vShort-circuiting stop; streamer state is already '%v' state", streamer.getLoglinePrefix(), streamer.state)
		return nil
	}

	if streamer.state != streaming {
		return stacktrace.NewError("Cannot stop streamer; streamer is not in 'streaming' state")
	}

	streamer.outputLogger.Tracef("%vSending signal to stop streaming thread...", streamer.getLoglinePrefix())

	//Closing the ReadCloser opened to prevent blocking
	if streamer.inputState == inputDocker {
		(streamer.inputReadCloser).Close()
	} else if streamer.inputState == inputFilePath {
		streamer.filePathStreamThreadShutdownChan <- true
	} else {
		return stacktrace.NewError("Detected unrecognized enum value '%v'; this is a bug in Kurtosis itself", streamer.inputState)
	}
	
	streamer.outputLogger.Tracef("%vSuccessfully sent signal to stop streaming thread", streamer.getLoglinePrefix())

	streamer.outputLogger.Tracef("%vWaiting until thread reports stopped, or %v timeout is hit...", streamer.getLoglinePrefix(), streamerStopTimeout)

	select {
	case <- streamer.streamThreadStoppedChan:
		streamer.outputLogger.Tracef("%vThread reported stop", streamer.getLoglinePrefix())
		streamer.state = terminated

		return nil
	case <- time.After(streamerStopTimeout):
		streamer.outputLogger.Tracef("%v%v timeout was hit", streamer.getLoglinePrefix(), streamerStopTimeout)
		streamer.state = failedToStop

		return stacktrace.NewError("We asked the streamer to stop but it still hadn't after %v", streamerStopTimeout)
	}

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (streamer *LogStreamer) startStreamingThread(input io.Reader, threadShutdownHook func()) error {

	if streamer.state != notStarted {
		return stacktrace.NewError("Cannot start streaming with this log streamer; streamer is not in the '%v' state", notStarted)
	}

	filePathStreamThreadShutdownChan := make(chan bool)
	streamThreadStoppedChan := make(chan bool)

	streamer.filePathStreamThreadShutdownChan = filePathStreamThreadShutdownChan
	streamer.streamThreadStoppedChan = streamThreadStoppedChan

	go func() {
		defer threadShutdownHook()

		if streamer.inputState == inputDocker {
			if _, err := stdcopy.StdCopy(streamer.outputLogger.Out, streamer.outputLogger.Out, input); err != nil {
				streamer.outputLogger.Errorf("%vAn error occurred reading the docker log output from the test logs: %v", streamer.getLoglinePrefix(), err)
			}
		} else if streamer.inputState == inputFilePath {
			streamFilePointerLogs(streamer, input)
		} else {
			streamer.outputLogger.Errorf("Detected unrecognized enum value '%v'; this is a bug in Kurtosis itself", streamer.inputState)
		}

		streamer.streamThreadStoppedChan <- true
		
	}()
	streamer.state = streaming

	return nil
}


func streamFilePointerLogs(streamer *LogStreamer, input io.Reader) error {

	keepGoing := true
	for keepGoing {
		streamer.outputLogger.Tracef("%vRunning channel-check cycle...", streamer.getLoglinePrefix())
		select {
		case <-streamer.filePathStreamThreadShutdownChan:
			streamer.outputLogger.Tracef("%vReceived signal on stream thread shutdown chan; setting keepGoing to false", streamer.getLoglinePrefix())
			keepGoing = false
		case <-time.After(timeBetweenStreamerCopies):
			streamer.outputLogger.Tracef("%vNo signal received on stream thread shutdown chan after waiting for %v; copying logs", streamer.getLoglinePrefix(), timeBetweenStreamerCopies)

			if _, err := io.Copy(streamer.outputLogger.Out, input); err != nil {
				streamer.outputLogger.Errorf("%vAn error occurred copying the output from the test logs: %v", streamer.getLoglinePrefix(), err)
			}
		}
		streamer.outputLogger.Tracef("%vChannel-check cycle completed", streamer.getLoglinePrefix())
	}
	// Do a final copy, to capture any non-copied output
	if _, err := io.Copy(streamer.outputLogger.Out, input); err != nil {
		streamer.outputLogger.Errorf("%vAn error occurred copying the final output from the test logs: %v", streamer.getLoglinePrefix(), err)
	}

	return nil
}


// This is a prefix that will be attached to log messages to identify that they're coming from the streamer
// This is necessary because these log messages will likely be outputted in the section labelled "testsuite logs",
// so we need to distinguish streamer logs (which come from the initializer) from logs that come from the testsuite
func (streamer LogStreamer) getLoglinePrefix() string {
	return fmt.Sprintf("[%v] ", streamer.loglineLabel)
}


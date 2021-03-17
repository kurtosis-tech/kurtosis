/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

// This class provides a switch that, when turned on, will capture any usages of the system-wide logger (e.g. 'logrus.Info')
// and store them for later use
type StdLoggerOutputCapturer struct {
	// Capture writer that will store any erroneous system logs while capturing is active
	interceptor             *erroneousSystemLogCaptureWriter
	writerBeforeManagement  io.Writer

	// Whether log messages written to logrus standard out are being intercepted or not
	isInterceptingStdLogger bool

	mutex *sync.Mutex
}

func newStdLoggerOutputCapturer() *StdLoggerOutputCapturer {
	return &StdLoggerOutputCapturer{
		interceptor:            	newErroneousSystemLogCaptureWriter(),
		isInterceptingStdLogger: 	false,
		writerBeforeManagement: 	nil,
		mutex:                   	&sync.Mutex{},
	}
}

/*
Starts intercepting any system-level logging for later display, rather than sending straight to STDOUT
*/
func (capturer *StdLoggerOutputCapturer) startInterceptingStdLogger() {
	capturer.mutex.Lock()
	defer capturer.mutex.Unlock()

	if capturer.isInterceptingStdLogger {
		return
	}

	stdLogger := logrus.StandardLogger()
	capturer.writerBeforeManagement = stdLogger.Out

	// No copy constructor :(
	capturer.sideChannelLogger = logrus.New()
	capturer.sideChannelLogger.SetOutput(stdLogger.Out)
	capturer.sideChannelLogger.SetFormatter(stdLogger.Formatter)
	capturer.sideChannelLogger.SetLevel(stdLogger.Level)
	// NOTE: we don't copy hooks here because we don't use them - if we ever use hooks, copy them here!

	logrus.SetOutput(capturer.interceptor)
	capturer.isInterceptingStdLogger = true
}

/*
Stops intercepting system-level logging
*/
func (capturer *StdLoggerOutputCapturer) stopInterceptingStdLogger() {
	capturer.mutex.Lock()
	capturer.mutex.Unlock()

	if !capturer.isInterceptingStdLogger {
		return
	}

	logrus.SetOutput(capturer.writerBeforeManagement)
	capturer.isInterceptingStdLogger = false
}

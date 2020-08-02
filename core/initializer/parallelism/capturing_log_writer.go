package parallelism

import (
	"runtime"
)

/*
Package struct encapsulating information about where an erroneous system logger event came from
 */
type ErroneousSystemLogInfo struct {
	Message    []byte
	Stacktrace []byte
}

/*
Because the tests will run in parallel and need to have their logs captured independently so they don't get all jumbled,
we expect the developer to write to test-specific loggers rather than the system logger. Developers might still forget,
however, so we need a way to:

1) loudly remind them in case they slip up and use the system-level logger but
2) not crash the program, because code we don't own can use the system logger (we tried panicking on
  system-level log write, but that didn't work because the Docker client writes to the system-level log)

Thus, we have this special writer that we plug in which doesn't actually write to STDOUT but captures the input for
 later logging in the form of a really loud error message.
 */
type ErroneousSystemLogCaptureWriter struct {
	logMessages []ErroneousSystemLogInfo
}

/*
Creates a new writer for capturing erroneous system log events
 */
func NewErroneousSystemLogCaptureWriter() *ErroneousSystemLogCaptureWriter {
	return &ErroneousSystemLogCaptureWriter{
		logMessages: []ErroneousSystemLogInfo{},
	}
}

/*
This write function (which comes from the Writer interface) will capture:
		a) the message that was intended for logging and
		b) the stacktrace at time of logging
	to make it easy for a developer to see where they're accidentally using the system-level log.
 */
func (writer *ErroneousSystemLogCaptureWriter) Write(data []byte) (n int, err error) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	stacktraceBytes := getStacktraceBytes()
	logInfo := ErroneousSystemLogInfo{
		Message:       dataCopy,
		Stacktrace: stacktraceBytes,
	}
	writer.logMessages = append(writer.logMessages, logInfo)
	return len(data), nil
}

/*
Retrieves the errenous system-level logger messages that were captured
 */
func (writer *ErroneousSystemLogCaptureWriter) GetCapturedMessages() []ErroneousSystemLogInfo {
	return writer.logMessages
}

/*
This code is almost an exact copy-paste from the stdlib's debug.PrintStack, because we need to have
	a buffer big enough to capture the stack trace... but we don't know in advance how big the stack trace
	will be.
 */
func getStacktraceBytes() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[0:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

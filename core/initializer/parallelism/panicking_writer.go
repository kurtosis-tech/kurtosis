package parallelism

/*
This is an implementation of io.Writer specially intended for the system-level logger while we expect the coder to be
writing to test-specific loggers, so that if the coder accidentally writes to the system-level logger (e.g.
logrus.Info rather than specificLogger.Info) they'll get a loud error
 */
type panickingLogWriter struct {}

func (writer panickingLogWriter) Write(p []byte) (n int, err error) {
	panic("The system-level logger was used in a spot where a test-specific logger should have been used instead; this is a code bug that needs to be corrected!")
}



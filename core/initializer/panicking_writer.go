package initializer

/*
This is a writer specially intended for the system-level logger while we expect the coder to be writing to test-specific loggers, so
that if the coder accidentally writes to the system-level logger (e.g. logrus.Info rather than specificLogger.Info) they'll get a loud error
 */
type PanickingLogWriter struct {}

func (writer PanickingLogWriter) Write(p []byte) (n int, err error) {
	panic("The system-level logger was called in a spot where a test-specific logger should have been called; this is a code bug that needs to be corrected!")
}



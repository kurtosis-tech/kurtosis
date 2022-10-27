package user_service_logs_read_closer

import "io"

const (
	lineBreakStr = "\n"
	userServiceLogLineChanBufferSize = 2
)

type UserServiceLogsReadCloser struct {
	userServiceLogLineChan chan string
}

func NewUserServiceLogsReadCloser() *UserServiceLogsReadCloser {
	userServiceLogLinesChan := make(chan string, userServiceLogLineChanBufferSize)
	return &UserServiceLogsReadCloser{userServiceLogLineChan: userServiceLogLinesChan}
}

func (readCloser *UserServiceLogsReadCloser) AddLine(newLine string)  {
	newLineWithLineBreak := newLine + lineBreakStr
	readCloser.userServiceLogLineChan <- newLineWithLineBreak
}

func (readCloser *UserServiceLogsReadCloser) Read(p []byte) (int, error) {
	userServiceLogLines, isChanOpen := <- readCloser.userServiceLogLineChan
	if !isChanOpen {
		return 0, io.EOF
	}
	n := copy(p, userServiceLogLines)
	return n, nil
}

func (readCloser *UserServiceLogsReadCloser) Close() error {
	close(readCloser.userServiceLogLineChan)
	return nil
}

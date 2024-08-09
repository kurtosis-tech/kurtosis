package logline

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"

const (
	batchLogsAmount    = 500
	logsChanBufferSize = 300
)

type LogLineSender struct {
	logsChan chan map[service.ServiceUUID][]LogLine

	logLineBuffer []LogLine
}

func NewLogLineSender() *LogLineSender {
	return &LogLineSender{
		logsChan:      make(chan map[service.ServiceUUID][]LogLine, logsChanBufferSize),
		logLineBuffer: []LogLine{},
	}
}

func (sender *LogLineSender) SendLogLine(serviceUuid service.ServiceUUID, logLine LogLine) {
	sender.logLineBuffer = append(sender.logLineBuffer, logLine)

	if len(sender.logLineBuffer)%batchLogsAmount == 0 {
		userServicesLogLinesMap := map[service.ServiceUUID][]LogLine{
			serviceUuid: sender.logLineBuffer,
		}
		sender.logsChan <- userServicesLogLinesMap

		// clear buffer after flushing it through the channel
		sender.logLineBuffer = []LogLine{}
	}
}

func (sender *LogLineSender) GetLogsChannel() chan map[service.ServiceUUID][]LogLine {
	return sender.logsChan
}

package logline

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"

const (
	batchLogsAmount = 500
)

type LogLineSender struct {
	logsChan chan map[service.ServiceUUID][]LogLine

	logLineBuffer []LogLine
}

func NewLogLineSender(logsChan chan map[service.ServiceUUID][]LogLine) *LogLineSender {
	return &LogLineSender{logsChan: logsChan}
}

func (sender *LogLineSender) SendLogLine(serviceUuid service.ServiceUUID, logLine LogLine) {
	sender.logLineBuffer = append(sender.logLineBuffer, logLine)

	if len(sender.logLineBuffer)%batchLogsAmount == 0 {
		userServicesLogLinesMap := map[service.ServiceUUID][]LogLine{
			serviceUuid: sender.logLineBuffer,
		}
		sender.logsChan <- userServicesLogLinesMap
		sender.logLineBuffer = []LogLine{}
	}
}

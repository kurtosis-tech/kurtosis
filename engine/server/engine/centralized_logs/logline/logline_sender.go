package logline

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"sync"
)

const (
	batchLogsAmount    = 1
	logsChanBufferSize = 1
)

type LogLineSender struct {
	logsChan chan map[service.ServiceUUID][]LogLine

	logLineBuffer map[service.ServiceUUID][]LogLine

	sync.Mutex
}

func NewLogLineSender() *LogLineSender {
	return &LogLineSender{
		logsChan:      make(chan map[service.ServiceUUID][]LogLine),
		logLineBuffer: map[service.ServiceUUID][]LogLine{},
	}
}

func (sender *LogLineSender) SendLogLine(serviceUuid service.ServiceUUID, logLine LogLine) {
	sender.Mutex.Lock()
	defer sender.Mutex.Unlock()

	sender.logLineBuffer[serviceUuid] = append(sender.logLineBuffer[serviceUuid], logLine)

	if len(sender.logLineBuffer[serviceUuid])%batchLogsAmount == 0 {
		userServicesLogLinesMap := map[service.ServiceUUID][]LogLine{
			serviceUuid: sender.logLineBuffer[serviceUuid],
		}
		sender.logsChan <- userServicesLogLinesMap

		// clear buffer after flushing it through the channel
		sender.logLineBuffer[serviceUuid] = []LogLine{}
	}
}

func (sender *LogLineSender) GetLogsChannel() chan map[service.ServiceUUID][]LogLine {
	return sender.logsChan
}

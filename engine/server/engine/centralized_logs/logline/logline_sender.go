package logline

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"sync"
)

const (
	batchLogsAmount    = 500
	logsChanBufferSize = 300
)

type LogLineSender struct {
	logsChan chan map[service.ServiceUUID][]LogLine

	logLineBuffer map[service.ServiceUUID][]LogLine

	mu sync.Mutex
}

func NewLogLineSender() *LogLineSender {
	return &LogLineSender{
		logsChan:      make(chan map[service.ServiceUUID][]LogLine, logsChanBufferSize),
		logLineBuffer: map[service.ServiceUUID][]LogLine{},
	}
}

func (sender *LogLineSender) SendLogLine(serviceUuid service.ServiceUUID, logLine LogLine) {
	sender.mu.Lock()
	defer sender.mu.Unlock()

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

// sends all logs remaining in the buffers through the channel
// this should be called at the end of processing to send the remainder of logs
func (sender *LogLineSender) Flush() {
	sender.mu.Lock()
	defer sender.mu.Unlock()

	for uuid, logLines := range sender.logLineBuffer {
		serviceUuid := uuid
		userServiceLogLinesMap := map[service.ServiceUUID][]LogLine{
			serviceUuid: logLines,
		}
		sender.logsChan <- userServiceLogLinesMap
	}
}

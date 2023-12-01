package to_http

import (
	"time"

	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"golang.org/x/exp/slices"

	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

func ToHttpServiceLogs(
	requestedServiceUuids []user_service.ServiceUUID,
	serviceLogsByServiceUuid map[user_service.ServiceUUID][]logline.LogLine,
	initialNotFoundServiceUuids []string,
) *api_type.ServiceLogs {
	serviceLogLinesByUuid := make(map[string]api_type.LogLine, len(serviceLogsByServiceUuid))
	notFoundServiceUuids := make([]string, len(initialNotFoundServiceUuids))
	for _, serviceUuid := range requestedServiceUuids {
		serviceUuidStr := string(serviceUuid)
		isInNotFoundUuidList := slices.Contains(initialNotFoundServiceUuids, serviceUuidStr)
		serviceLogLines, found := serviceLogsByServiceUuid[serviceUuid]
		// should continue in the not-found-UUID list
		if !found && isInNotFoundUuidList {
			notFoundServiceUuids = append(notFoundServiceUuids, serviceUuidStr)
		}

		// there is no new log lines but is a found UUID, so it has to be included in the service logs map
		if !found && !isInNotFoundUuidList {
			serviceLogLinesByUuid[serviceUuidStr] = api_type.LogLine{
				Line:      []string{},
				Timestamp: time.Now(),
			}
		}

		logLines := ToHttpLogLines(serviceLogLines)
		serviceLogLinesByUuid[serviceUuidStr] = logLines
	}

	response := &api_type.ServiceLogs{
		NotFoundServiceUuidSet:   &notFoundServiceUuids,
		ServiceLogsByServiceUuid: &serviceLogLinesByUuid,
	}
	return response
}

func ToHttpLogLines(logLines []logline.LogLine) api_type.LogLine {
	logLinesStr := make([]string, len(logLines))
	var logTimestamp time.Time

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
		logTimestamp = logLine.GetTimestamp()
	}

	return api_type.LogLine{Line: logLinesStr, Timestamp: logTimestamp}

}

package kurtosis_context

import "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"

//This struct wrap the information returned by the user service logs GRPC stream
type serviceLogsStreamContent struct{
	serviceLogsByServiceGuids map[services.ServiceGUID][]*ServiceLog
	notFoundServiceGuids map[services.ServiceGUID]bool
}

func newServiceLogsStreamContent(
	serviceLogsByServiceGuids map[services.ServiceGUID][]*ServiceLog,
	notFoundServiceGuids map[services.ServiceGUID]bool,
) *serviceLogsStreamContent {
	return &serviceLogsStreamContent{
		serviceLogsByServiceGuids: serviceLogsByServiceGuids,
		notFoundServiceGuids: notFoundServiceGuids,
	}
}

func (streamContent *serviceLogsStreamContent) GetServiceLogsByServiceGuids() map[services.ServiceGUID][]*ServiceLog {
	return streamContent.serviceLogsByServiceGuids
}

func (streamContent *serviceLogsStreamContent) GetNotFoundServiceGuids() map[services.ServiceGUID]bool {
	return streamContent.notFoundServiceGuids
}

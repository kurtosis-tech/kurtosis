package kurtosis_context

import "github.com/kurtosis-tech/kurtosis/api/golang/v2/v2/core/lib/services"

// This struct wrap the information returned by the user service logs GRPC stream
type serviceLogsStreamContent struct {
	serviceLogsByServiceUuids map[services.ServiceUUID][]*ServiceLog
	notFoundServiceUuids      map[services.ServiceUUID]bool
}

func newServiceLogsStreamContent(
	serviceLogsByServiceUuids map[services.ServiceUUID][]*ServiceLog,
	notFoundServiceUuids map[services.ServiceUUID]bool,
) *serviceLogsStreamContent {
	return &serviceLogsStreamContent{
		serviceLogsByServiceUuids: serviceLogsByServiceUuids,
		notFoundServiceUuids:      notFoundServiceUuids,
	}
}

// Docs available at https://docs.kurtosis.com/sdk#getservicelogsbyserviceuuids----mapserviceuuid-arrayservicelog-servicelogsbyserviceuuids
func (streamContent *serviceLogsStreamContent) GetServiceLogsByServiceUuids() map[services.ServiceUUID][]*ServiceLog {
	return streamContent.serviceLogsByServiceUuids
}

// Docs available at https://docs.kurtosis.com/sdk#getnotfoundserviceuuids---setserviceuuid-notfoundserviceuuids
func (streamContent *serviceLogsStreamContent) GetNotFoundServiceUuids() map[services.ServiceUUID]bool {
	return streamContent.notFoundServiceUuids
}

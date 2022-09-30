package centralized_logs

import (
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"io"
	"io/ioutil"
	"strings"
)

const (
	minimalKurtosisBackendMockTestUserService1Guid = "test-user-service-1"
	minimalKurtosisBackendMockTestUserService2Guid = "test-user-service-2"
	minimalKurtosisBackendMockTestUserService3Guid = "test-user-service-3"

	minimalKurtosisBackendMockTestUserService1LogLines = "This is the first user service #1 log line.\nThis is the second one.\nThis is the third one."
	minimalKurtosisBackendMockTestUserService2LogLines = "This is the first user service #2 log line.\nThis is the second one.\nThis is the third one."
	minimalKurtosisBackendMockTestUserService3LogLines = "This is the first user service #3 log line.\nThis is the second one.\nThis is the third one."

	minimalKurtosisBackendMockFakeErrorStr = "fake error"
)


//TODO replace with the final KurtosisBackend's bock when we have it
//This is a temporary hack until we get the real MockedKurtosisBackend created by Mockery
type MinimalKurtosisBackendMock struct{}

func NewMinimalKurtosisBackendMock() *MinimalKurtosisBackendMock {
	return &MinimalKurtosisBackendMock{}
}

func (mock MinimalKurtosisBackendMock) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	var (
		successfulUserServiceLogLinesByGuids = map[service.ServiceGUID]string{
			minimalKurtosisBackendMockTestUserService1Guid: minimalKurtosisBackendMockTestUserService1LogLines,
			minimalKurtosisBackendMockTestUserService2Guid: minimalKurtosisBackendMockTestUserService2LogLines,
			minimalKurtosisBackendMockTestUserService3Guid: minimalKurtosisBackendMockTestUserService3LogLines,
		}

		fakeError = errors.New(minimalKurtosisBackendMockFakeErrorStr)
	)

	successfulUserServiceLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}

	for userServiceGuid := range filters.GUIDs {

		if userServiceLogsLinesStr, found := successfulUserServiceLogLinesByGuids[userServiceGuid]; found {
			logLinesReader := strings.NewReader(userServiceLogsLinesStr)
			logLinesReadCloser := ioutil.NopCloser(logLinesReader)

			successfulUserServiceLogs[userServiceGuid] = logLinesReadCloser
			continue
		}

		erroredUserServices[userServiceGuid] = fakeError

	}

	return successfulUserServiceLogs, erroredUserServices, nil
}

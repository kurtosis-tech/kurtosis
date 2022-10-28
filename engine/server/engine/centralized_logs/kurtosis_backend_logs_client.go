package centralized_logs

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

const (
	getUserServiceLogsShouldFollowLogsOption    = false
	streamUserServiceLogsShouldFollowLogsOption = true

	logsByKurtosisUserServiceGuidChanBufferSize = 2
	errorChanBufferSize                         = 2
)

type kurtosisBackendLogClient struct {
	//TODO temporary hack replace it with backend_interface.KurtosisBackend when we add the KurtosisBackendMock with Mockery
	kurtosisBackend MinimalKurtosisBackend
}

func NewKurtosisBackendLogClient(kurtosisBackend MinimalKurtosisBackend) *kurtosisBackendLogClient {
	return &kurtosisBackendLogClient{kurtosisBackend: kurtosisBackend}
}

func (client *kurtosisBackendLogClient) GetUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (map[service.ServiceGUID][]string, error) {

	resultLogsByKurtosisUserServiceGuid := map[service.ServiceGUID][]string{}

	userServiceFilters := &service.ServiceFilters{
		GUIDs: userServiceGuids,
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := client.kurtosisBackend.GetUserServiceLogs(ctx, enclaveID, userServiceFilters, getUserServiceLogsShouldFollowLogsOption)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}
	//Closing all the received read closers, this is in a separate loop because using the "defer"
	//instruction in a loop could cause possible resources leak
	defer func() {
		for _, userServiceReadCloserLogs := range successfulUserServiceLogs {
			userServiceReadCloserLogs.Close()
		}
	}()

	if len(erroredUserServiceGuids) > 0 {
		errorsFoundInServices := []string{}
		for userServiceGuid, errorInService := range erroredUserServiceGuids {
			errorFoundInServiceStr := fmt.Sprintf("Error in service with GUID '%v': %v", userServiceGuid, errorInService)
			errorsFoundInServices = append(errorsFoundInServices, errorFoundInServiceStr)
		}
		errorsFoundStr := strings.Join(errorsFoundInServices, "\n")

		return nil, stacktrace.NewError("Some user services returned with error when calling for the logs using filters '%+v'. Errors returned: \n%v", userServiceFilters, errorsFoundStr)
	}

	for userServiceGuid, userServiceReadCloserLogs := range successfulUserServiceLogs {
		userServiceLogsScanner := bufio.NewScanner(userServiceReadCloserLogs)
		userServiceLogsScanner.Split(bufio.ScanLines)

		userServiceLogsLines := []string{}
		for userServiceLogsScanner.Scan() {
			userServiceLogsLines = append(userServiceLogsLines, userServiceLogsScanner.Text())
		}

		resultLogsByKurtosisUserServiceGuid[userServiceGuid] = userServiceLogsLines
	}

	return resultLogsByKurtosisUserServiceGuid, nil
}

func (client *kurtosisBackendLogClient) StreamUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (
	chan map[service.ServiceGUID][]string,
	chan error,
	error,
) {
	logsByKurtosisUserServiceGuidChan := make(chan map[service.ServiceGUID][]string, logsByKurtosisUserServiceGuidChanBufferSize)
	errChan := make(chan error, errorChanBufferSize)

	userServiceFilters := &service.ServiceFilters{
		GUIDs: userServiceGuids,
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := client.kurtosisBackend.GetUserServiceLogs(ctx, enclaveID, userServiceFilters, streamUserServiceLogsShouldFollowLogsOption)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}
	//We can't directly defer the read closer's close calls here because we need them open, this is a stream flow

	closeUserServiceReadClosersFunc := func() {
		for _, userServiceReadCloserLogs := range successfulUserServiceLogs {
			userServiceReadCloserLogs.Close()
		}
	}

	if len(erroredUserServiceGuids) > 0 {
		errorsFoundInServices := []string{}
		for userServiceGuid, errorInService := range erroredUserServiceGuids {
			errorFoundInServiceStr := fmt.Sprintf("Error in service with GUID '%v': %v", userServiceGuid, errorInService)
			errorsFoundInServices = append(errorsFoundInServices, errorFoundInServiceStr)
		}
		errorsFoundStr := strings.Join(errorsFoundInServices, "\n")

		closeUserServiceReadClosersFunc()
		return nil, nil, stacktrace.NewError("Some user services returned with error when calling for the logs using filters '%+v'. Errors returned: \n%v", userServiceFilters, errorsFoundStr)
	}

	for successfulUserServiceGuid, successfulUserServiceReadCloserLogs := range successfulUserServiceLogs {
		go scanUserServiceLogsAndAddLogLinesToTheUserServiceLogsChan(
			ctx,
			successfulUserServiceGuid,
			successfulUserServiceReadCloserLogs,
			logsByKurtosisUserServiceGuidChan,
			errChan,
		)
	}

	return logsByKurtosisUserServiceGuidChan, errChan, nil
}

func (client *kurtosisBackendLogClient) FilterExistingServiceGuids(ctx context.Context, enclaveId enclave.EnclaveID, userServiceGuids map[service.ServiceGUID]bool) (map[service.ServiceGUID]bool, error) {
	logrus.Warnf("Unable to get the exhaustive list of service GUIDs for this backend. Original list of service GUIDs will be returned.")
	return userServiceGuids, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func scanUserServiceLogsAndAddLogLinesToTheUserServiceLogsChan(
	ctx context.Context,
	userServiceGUID service.ServiceGUID,
	userServiceReadCloser io.ReadCloser,
	logsByKurtosisUserServiceGuidChan chan map[service.ServiceGUID][]string,
	errChan chan error,
) {

	defer func() {
		if err := userServiceReadCloser.Close(); err != nil {
			logrus.Errorf("Streaming user service has finished, so we tried to close the user service read closer for service with GUID '%v', but an error was thrown:\n%v", userServiceGUID, err)
		}
	}()

	userServiceLogsScanner := bufio.NewScanner(userServiceReadCloser)
	userServiceLogsScanner.Split(bufio.ScanLines)

	for {
		select {
		case <-ctx.Done():
			errChan <- stacktrace.Propagate(ctx.Err(), "An error occurred streaming user service logs from Kurtosis backend logs client, the request context has done")
			return
		default:
			for userServiceLogsScanner.Scan() {
				newUserServiceLogLine := map[service.ServiceGUID][]string{
					userServiceGUID: {userServiceLogsScanner.Text()},
				}
				logsByKurtosisUserServiceGuidChan <- newUserServiceLogLine
			}
		}
	}

}

package centralized_logs

import (
	"bufio"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	shouldFollowLogs = false
)

type KurtosisBackendLogClient struct {
	//TODO temporary hack replace it with backend_interface.KurtosisBackend when we add the KurtosisBackendMock with Mockery
	kurtosisBackend MinimalKurtosisBackend
}

func NewKurtosisBackendLogClient(kurtosisBackend MinimalKurtosisBackend) *KurtosisBackendLogClient {
	return &KurtosisBackendLogClient{kurtosisBackend: kurtosisBackend}
}

func (client *KurtosisBackendLogClient) GetUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (map[service.ServiceGUID][]string, error) {

	resultLogsByKurtosisUserServiceGuid := map[service.ServiceGUID][]string{}

	userServiceFilters := &service.ServiceFilters{
		GUIDs: userServiceGuids,
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := client.kurtosisBackend.GetUserServiceLogs(ctx, enclaveID, userServiceFilters, shouldFollowLogs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}
	//Closing all the received read closers, this is in a separate loop because using the "defer"
	//instruction in a loop could cause possible resources leak
	defer func() {
		for _, userServiceReadCloserLogs := range successfulUserServiceLogs{
			userServiceReadCloserLogs.Close()
		}
	}()

	if len(erroredUserServiceGuids) > 0 {
		errorsFoundInServices := []string{}
		for userServiceGuid, errorInService := range erroredUserServiceGuids {
			errorFoundInServiceStr := fmt.Sprintf("Error in service with GUID '%v': %v", userServiceGuid, errorInService )
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

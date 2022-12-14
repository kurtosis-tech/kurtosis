//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because, as of 2022-10-28, the centralized logs feature is not implemented in Kubernetes yet
//TODO remove this comments after Kubernetes implementation

package search_logs_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
	"time"

	//"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "search-logs"
	isPartitioningEnabled = false

	dockerGettingStartedImage = "docker/getting-started"
	exampleServiceIdPrefix    = "search-logs-"

	shouldNotFollowLogs = false

	service1ServiceID services.ServiceID = exampleServiceIdPrefix + "service-1"

	firstFilterText     = "Starting feature"
	secondFilterText    = "network"
	matchRegexFilterStr = "Starting.*logs'"

	waitForAllLogsBeingCollectedInSeconds = 3

	testTimeOut = 90 * time.Second

	logLine1 = "Starting feature 'centralized logs'"
	logLine2 = "Starting feature 'network partitioning'"
	logLine3 = "Starting feature 'network soft partitioning'"
	logLine4 = "The data have being loaded"
)

var (
	expectedNonExistenceServiceGuids = map[services.ServiceGUID]bool{}

	service1LogLines = []string{
		logLine1,
		logLine2,
		logLine3,
		logLine4,
	}

	logLinesByService = map[services.ServiceID][]string{
		service1ServiceID: service1LogLines,
	}

	doesContainTextFilter          = kurtosis_context.NewDoesContainTextLogLineFilter(firstFilterText)
	doesNotContainTextFilter       = kurtosis_context.NewDoesNotContainTextLogLineFilter(secondFilterText)
	doesContainMatchRegexFilter    = kurtosis_context.NewDoesContainMatchRegexLogLineFilter(matchRegexFilterStr)
	doesNotContainMatchRegexFilter = kurtosis_context.NewDoesNotContainMatchRegexLogLineFilter(matchRegexFilterStr)

	filtersByRequest = []kurtosis_context.LogLineFilter{
		*doesContainTextFilter,
		*doesNotContainTextFilter,
		*doesContainMatchRegexFilter,
		*doesNotContainMatchRegexFilter,
	}

	expectedLogLinesByRequest = [][]string{
		{
			logLine1,
			logLine2,
			logLine3,
		},
		{
			logLine1,
			logLine4,
		},
		{
			logLine1,
		},
		{
			logLine2,
			logLine3,
			logLine4,
		},
	}

)

func TestSearchLogs(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	serviceList := addServices(t, enclaveCtx, logLinesByService)
	require.Equal(t, len(logLinesByService), len(serviceList))

	enclaveId := enclaveCtx.GetEnclaveID()

	userServiceGuids := map[services.ServiceGUID]bool{}
	for _, serviceCtx := range serviceList {
		serviceGuid := serviceCtx.GetServiceGUID()
		userServiceGuids[serviceGuid] = true
	}

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	for requestIndex, filter := range filtersByRequest {
		testEvaluationErr, receivedLogLines, receivedNotFoundServiceGuids := getLogsResponse(
			t,
			ctx,
			kurtosisCtx,
			enclaveId,
			userServiceGuids,
			&filter,
		)

		require.NoError(t, testEvaluationErr)
		require.Equal(t, expectedLogLinesByRequest[requestIndex], receivedLogLines)
		require.Equal(t, expectedNonExistenceServiceGuids, receivedNotFoundServiceGuids)
	}
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getLogsResponse(
	t *testing.T,
	ctx context.Context,
	kurtosisCtx *kurtosis_context.KurtosisContext,
	enclaveId enclaves.EnclaveID,
	userServiceGuids map[services.ServiceGUID]bool,
	logLineFilter *kurtosis_context.LogLineFilter,
) (
	error,
	[]string,
	map[services.ServiceGUID]bool,
) {
	serviceLogsStreamContentChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, enclaveId, userServiceGuids, shouldNotFollowLogs, logLineFilter)
	require.NoError(t, err, "An error occurred getting user service logs from user services with GUIDs '%+v' in enclave '%v' and with follow logs value '%v'", userServiceGuids, enclaveId, shouldNotFollowLogs)
	defer cancelStreamUserServiceLogsFunc()

	receivedNotFoundServiceGuids := map[services.ServiceGUID]bool{}
	receivedLogLines := []string{}

	var testEvaluationErr error

	shouldContinueInTheLoop := true

	for shouldContinueInTheLoop {
		select {
		case <-time.Tick(testTimeOut):
			testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
			shouldContinueInTheLoop = false
			break
		case serviceLogsStreamContent, isChanOpen := <-serviceLogsStreamContentChan:
			if !isChanOpen {
				shouldContinueInTheLoop = false
				break
			}

			serviceLogsByGuid := serviceLogsStreamContent.GetServiceLogsByServiceGuids()
			receivedNotFoundServiceGuids = serviceLogsStreamContent.GetNotFoundServiceGuids()

			for _, serviceLogLines := range serviceLogsByGuid {
				for _, serviceLogLine := range serviceLogLines {
					receivedLogLines = append(receivedLogLines, serviceLogLine.GetContent())
				}
			}
		}
	}

	return testEvaluationErr, receivedLogLines, receivedNotFoundServiceGuids
}

func addServices(
	t *testing.T,
	enclaveCtx *enclaves.EnclaveContext,
	logLinesByServiceID map[services.ServiceID][]string,
) map[services.ServiceID]*services.ServiceContext {

	servicesAdded := make(map[services.ServiceID]*services.ServiceContext, len(logLinesByServiceID))
	for serviceId, logLines := range logLinesByServiceID {
		containerConfig := getExampleServiceConfig(logLines)
		serviceCtx, err := enclaveCtx.AddService(serviceId, containerConfig)
		require.NoError(t, err, "An error occurred adding service with ID %v", serviceId)
		servicesAdded[serviceId] = serviceCtx
	}
	return servicesAdded
}

func getExampleServiceConfig(logLines []string) *services.ContainerConfig {

	entrypointArgs := []string{"/bin/sh", "-c"}

	var logLinesWithQuotes []string
	for _, logLine := range logLines {
		logLineWithQuote := fmt.Sprintf("\"%s\"", logLine)
		logLinesWithQuotes = append(logLinesWithQuotes, logLineWithQuote)
	}

	logLineSeparator := " "
	logLinesStr := strings.Join(logLinesWithQuotes, logLineSeparator)
	echoLogLinesLoopCmdStr := fmt.Sprintf("for i in %s; do echo \"$i\"; done;", logLinesStr)

	cmdArgs := []string{echoLogLinesLoopCmdStr}

	containerConfig := services.NewContainerConfigBuilder(
		dockerGettingStartedImage,
	).WithEntrypointOverride(
		entrypointArgs,
	).WithCmdOverride(
		cmdArgs,
	).Build()
	return containerConfig
}

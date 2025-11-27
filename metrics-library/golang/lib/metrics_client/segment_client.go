package metrics_client

import (
	"runtime"
	"strconv"
	"time"

	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/event"
	metrics_source "github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/segmentio/backo-go"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	//This key was generated in Kurtosis' Segment account for the `Kurtosis-Metrics-Library` source
	//You can configure this source on this page: https://app.segment.com/kurtosis/sources/kurtosis-metrics-library/overview
	accountWriteKey = "KpA8kDssJU1j0kuBZ0r2A81wuD1yisOn"

	shouldTrackIdentifyUserEventWhenClientIsCreated = false

	segmentClientFlushInterval = 10 * time.Minute

	retryBackoffBaseDuration = time.Second * 5
	retryBackoffFactor       = 3
	retryBackoffJitter       = 0
	retryBackoffCap          = time.Hour * 24

	batchSizeValueForFlushAfterEveryEvent = 1

	isCIKey            = "is_ci"
	osKey              = "os"
	archKey            = "arch"
	backendKey         = "backend"
	cloudUserIdKey     = "cloud_user_id"
	cloudInstanceIdKey = "cloud_instance_id"
	isCloudKey         = "is_cloud"
)

type segmentClient struct {
	client           analytics.Client
	analyticsContext *analytics.Context
	userID           string
	isCI             string
	backendType      string
	cloudUserId      CloudUserID
	cloudInstanceId  CloudInstanceID
}

// The argument shouldFlushQueueOnEachEvent is used to imitate a sync request, it is not exactly the same because
// the event is enqueued but the queue is flushed suddenly so is pretty close to event traked in sync
// The argument callbackObject is an object that will be used by the client to notify the
// application when messages sends to the backend API succeeded or failed.
func newSegmentClient(source metrics_source.Source, sourceVersion string, userId string, backendType string, shouldFlushQueueOnEachEvent bool, callbackObject analytics.Callback, logger analytics.Logger, isCI bool, cloudUserId CloudUserID, cloudInstanceId CloudInstanceID) (*segmentClient, error) {

	// nolint: exhaustruct
	config := analytics.Config{
		//The flushing interval of the client
		Interval: segmentClientFlushInterval,
		//NOTE: Segment client has a max attempt = 10, so this retry strategy
		//allow us to execute the first attempt in 5 seconds and the last attend in 24 hours
		//which is useful if a user is executing the metrics without internet connection for several hours
		RetryAfter: func(attempt int) time.Duration {
			retryBacko := backo.NewBacko(retryBackoffBaseDuration, retryBackoffFactor, retryBackoffJitter, retryBackoffCap)
			return retryBacko.Duration(attempt)
		},
		Callback: callbackObject,

		Logger: logger,
	}

	if shouldFlushQueueOnEachEvent {
		//if BatchSize is equal = 1 the event will being send immediately
		config.BatchSize = batchSizeValueForFlushAfterEveryEvent
	}

	client, err := analytics.NewWithConfig(accountWriteKey, config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new Segment client with config '%+v'", config)
	}

	analyticsContext := newAnalyticsContext(source, sourceVersion)

	//We could activate this functionality if we want to track an event to identify the user
	//every time the client is created
	if shouldTrackIdentifyUserEventWhenClientIsCreated {
		// nolint: exhaustruct
		if err := client.Enqueue(analytics.Identify{
			UserId:  userId,
			Context: analyticsContext,
		}); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred enqueuing a new identify event in Segment client's queue")
		}
	}

	return &segmentClient{client: client, analyticsContext: analyticsContext, userID: userId, isCI: strconv.FormatBool(isCI), backendType: backendType, cloudUserId: cloudUserId, cloudInstanceId: cloudInstanceId}, nil
}

func (segment *segmentClient) TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics bool) error {
	newEvent := event.NewShouldSendMetricsUserElectionEvent(didUserAcceptSendingMetrics)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking should-send-metrics user election")
	}

	return nil
}

func (segment *segmentClient) TrackUserSharedEmailAddress(userSharedEmailAddress string) error {
	newEvent := event.NewUserSharesEmailAddress(userSharedEmailAddress)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking user-shares-email-address event")
	}
	return nil
}

func (segment *segmentClient) TrackCreateEnclave(enclaveId string, isSubnetworkingEnabled bool) error {
	newEvent := event.NewCreateEnclaveEvent(enclaveId, isSubnetworkingEnabled)

	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking create enclave event")
	}
	return nil
}

func (segment *segmentClient) TrackStopEnclave(enclaveId string) error {
	newEvent := event.NewStopEnclaveEvent(enclaveId)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking stop enclave event")
	}
	return nil
}

func (segment *segmentClient) TrackDestroyEnclave(enclaveId string) error {
	newEvent := event.NewDestroyEnclaveEvent(enclaveId)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking destroy enclave event")
	}
	return nil
}

func (segment *segmentClient) TrackKurtosisRun(packageId string, isRemote bool, isDryRun bool, isScript bool, serializedParams string) error {
	newEvent := event.NewKurtosisRunEvent(packageId, isRemote, isDryRun, isScript, serializedParams)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking run kurtosis event")
	}
	return nil
}

func (segment *segmentClient) TrackServiceUpdate(enclaveId string, serviceId string) error {
	newEvent := event.NewUpdateServiceEvent(enclaveId, serviceId)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking service update event")
	}
	return nil
}

func (segment *segmentClient) TrackStartService(enclaveId string, serviceId string) error {
	newEvent := event.NewStartServiceEvent(enclaveId, serviceId)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking start service event")
	}
	return nil
}

func (segment *segmentClient) TrackStopService(enclaveId string, serviceId string) error {
	newEvent := event.NewStopServiceEvent(enclaveId, serviceId)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking stop service event")
	}
	return nil
}

func (segment *segmentClient) TrackKurtosisRunFinishedEvent(packageId string, numberOfServices int, isSuccess bool, serializedParams string) error {
	newEvent := event.NewKurtosisRunFinishedEvent(packageId, numberOfServices, isSuccess, serializedParams)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking kurtosis run finished event")
	}
	return nil
}

func (segment *segmentClient) TrackKurtosisAnalyticsToggle(analyticsStatus bool) error {
	newEvent := event.NewKurtosisAnalyticsToggleEvent(analyticsStatus)
	if err := segment.track(newEvent); err != nil {
		return stacktrace.Propagate(err, "an error occurred while tracking kurtosis analytics toggle event")
	}
	return nil
}

func (segment *segmentClient) close() (err error) {
	if err := segment.client.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the Segment client")
	}
	return nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func (segment *segmentClient) track(event *event.Event) error {

	propertiesToTrack := analytics.NewProperties()

	eventProperties := event.GetProperties()

	for propertyKey, propertyValue := range eventProperties {
		propertiesToTrack.Set(propertyKey, propertyValue)
	}

	propertiesToTrack.Set(isCIKey, segment.isCI)
	propertiesToTrack.Set(osKey, runtime.GOOS)
	propertiesToTrack.Set(archKey, runtime.GOARCH)
	propertiesToTrack.Set(backendKey, segment.backendType)
	propertiesToTrack.Set(cloudUserIdKey, segment.cloudUserId)
	propertiesToTrack.Set(cloudInstanceIdKey, segment.cloudInstanceId)
	propertiesToTrack.Set(isCloudKey, segment.cloudInstanceId != "" || segment.cloudUserId != "")

	// nolint: exhaustruct
	if err := segment.client.Enqueue(analytics.Track{
		Event:      event.GetName(),
		UserId:     segment.userID,
		Context:    segment.analyticsContext,
		Properties: propertiesToTrack,
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred enqueuing a new event with name '%v' and properties '%+v' in Segment client's queue", event.GetName(), propertiesToTrack)
	}
	return nil
}

func newAnalyticsContext(source metrics_source.Source, sourceVersion string) *analytics.Context {
	// nolint: exhaustruct
	appInfo := analytics.AppInfo{
		Name:    source.GetKey(),
		Version: sourceVersion,
	}

	// nolint: exhaustruct
	analyticsContext := &analytics.Context{App: appInfo}

	return analyticsContext
}

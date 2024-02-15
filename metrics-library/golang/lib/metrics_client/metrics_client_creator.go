package metrics_client

import (
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultMetricsType = Segment
)

// The argument shouldFlushQueueOnEachEvent is used to imitate a sync request, it is not exactly the same because
// the event is enqueued but the queue is flushed suddenly so is pretty close to event traked in sync
// The argument callbackObject is an object that will be used by the client to notify the
// application when messages sends to the backend API succeeded or failed.
func CreateMetricsClient(options *CreateMetricsClientOption) (MetricsClient, func() error, error) {

	metricsClientType := DoNothing

	if options.didUserAcceptSendingMetrics {
		metricsClientType = defaultMetricsType
	}

	switch metricsClientType {
	case Segment:
		segmentCallback := newSegmentCallback(options.callbackObject.Success, options.callbackObject.Failure)
		metricsClient, err := newSegmentClient(options.source, options.sourceVersion, options.userId, options.backendType, options.shouldFlushQueueOnEachEvent, segmentCallback, options.logger, options.isCI, options.cloudUserId, options.cloudInstanceId)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred creating Segment metrics client")
		}
		return metricsClient, metricsClient.close, nil
	case DoNothing:
		metricsClient := newDoNothingClient(options.callbackObject)
		return metricsClient, metricsClient.close, nil
	default:
		return nil, nil, stacktrace.NewError("Unrecognized metrics client type '%v'", metricsClientType)
	}
}

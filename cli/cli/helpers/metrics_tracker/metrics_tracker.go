package metrics_tracker

import (
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/event"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	yesStr = "yes"
	noStr = "no"
)

type MetricsTracker struct {
	client client.MetricsClient
}

func NewMetricsTracker(client client.MetricsClient) *MetricsTracker {
	return &MetricsTracker{client: client}
}

func (tracker *MetricsTracker) TrackUserAcceptSendingMetrics(userAcceptSendingMetrics bool) error {

	var metricsLabel string
	if userAcceptSendingMetrics{
		metricsLabel = yesStr
	} else {
		metricsLabel = noStr
	}

	metricsEvent, err := event.NewEventBuilder(event.InstallCategory, event.ConsentAction).
		WithLabel(metricsLabel).
		Build()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new metrics event")
	}

	if err := tracker.client.Track(metricsEvent); err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking metrics event &+v", metricsEvent)
	}

	return nil
}

func (tracker *MetricsTracker) DisableTracking() {
	tracker.DisableTracking()
}

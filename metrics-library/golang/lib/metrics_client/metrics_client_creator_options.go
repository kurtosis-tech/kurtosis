package metrics_client

import (
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/source"
	"gopkg.in/segmentio/analytics-go.v3"
)

type CreateMetricsClientOption struct {
	source                      source.Source
	sourceVersion               string
	userId                      string
	backendType                 string
	didUserAcceptSendingMetrics bool
	shouldFlushQueueOnEachEvent bool
	callbackObject              Callback
	logger                      analytics.Logger
	isCI                        bool
}

func NewMetricsClientCreatorOption(source source.Source,
	sourceVersion string,
	userId string,
	backendType string,
	didUserAcceptSendingMetrics bool,
	shouldFlushQueueOnEachEvent bool,
	callbackObject Callback,
	logger analytics.Logger,
	isCI bool) *CreateMetricsClientOption {
	return &CreateMetricsClientOption{
		source:                      source,
		sourceVersion:               sourceVersion,
		userId:                      userId,
		backendType:                 backendType,
		didUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		shouldFlushQueueOnEachEvent: shouldFlushQueueOnEachEvent,
		callbackObject:              callbackObject,
		logger:                      logger,
		isCI:                        isCI,
	}
}

package metrics_client

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type segmentCallback struct {
	successFunc func()
	failureFunc func(error)
}

func newSegmentCallback(successFunc func(), failureFunc func(error)) *segmentCallback {
	return &segmentCallback{successFunc: successFunc, failureFunc: failureFunc}
}

func (callback *segmentCallback) Success(msg analytics.Message) {
	logrus.Debugf("Metrics Client success callback executed with message '%+v'", msg)
	callback.successFunc()
}

func (callback *segmentCallback) Failure(msg analytics.Message, err error) {
	logrus.Debugf("Metrics client failure callback executed with message '%+v' and error %v", msg, err)
	callback.failureFunc(err)
}

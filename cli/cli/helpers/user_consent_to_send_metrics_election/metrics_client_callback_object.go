package user_consent_to_send_metrics_election

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

type metricsClientCallbackObject struct{}

func (metricsClientCallbackObject *metricsClientCallbackObject) Success(msg analytics.Message) {
	//If the metrics was sent successfully the file will be removed because it indicates that
	//the system doesn't have to send it anymore
	logrus.Debugf("Metrics Client success callback executed with message '%+v'", msg)
	userConsentToSendMetricsElectionStore := GetUserConsentToSendMetricsElectionStore()
	if err := userConsentToSendMetricsElectionStore.remove(); err != nil {
		//We do nothing when removing the user consent to send metrics election file fails, it will
		//be retried next time users execute engine start command
		logrus.Debugf("We tried remove the user consent to send metrics election file, but doing so threw an error:\n%v", err)
	}
}

func (metricsClientCallbackObject *metricsClientCallbackObject) Failure(msg analytics.Message, err error) {
	//We do nothing when sending metrics consent request fails, it will be retried next time
	//users execute engine start command
	logrus.Debugf("Metrics client failure callback executed with message '%+v' and error %v", msg, err)
}

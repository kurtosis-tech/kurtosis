package user_send_metrics_election

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/user_send_metrics_election/file_backed_user_send_metrics_election_event_backlog"
	"github.com/sirupsen/logrus"
)

type metricsElectionEventBacklogClearingCallback struct{}

func (metricsClientCallbackObject *metricsElectionEventBacklogClearingCallback) Success() {
	filepath, err := host_machine_directories.GetUserSendMetricsElectionFilepath()
	if err != nil {
		logrus.Debugf( "An error occurred getting the user-send-metrics-election filepath\n%v", err)
	}
	//If the metrics was sent successfully the file will be removed because it indicates that
	//the system doesn't have to send it anymore
	fileBackedUserSendMetricsElectionEventBacklog := file_backed_user_send_metrics_election_event_backlog.GetFileBackedUserSendMetricsElectionEventBacklog(filepath)
	if err := fileBackedUserSendMetricsElectionEventBacklog.Clear(); err != nil {
		//We do nothing when removing the user consent to send metrics election file fails, it will
		//be retried next time users execute engine start command
		logrus.Debugf("We tried to clear the user-send-metrics-election-event-backlog, but doing so threw an error:\n%v", err)
	}
}

func (metricsClientCallbackObject *metricsElectionEventBacklogClearingCallback) Failure(err error) {
	//We do nothing when sending metrics consent request fails, it will be retried next time
	//users execute engine start command
	logrus.Debugf("Metrics client failure callback executed with error: %v",  err)
}

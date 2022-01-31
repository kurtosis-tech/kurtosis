package kurtosis_config

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_optin"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/prompt_displayer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	metricsPromptLabel = "Is it okay to send anonymized metrics purely to improve the product?"
	metricsConsentPromptLabel = "Ok you do not want to send metrics, but is it okay to send only that you reject sending metrics?"

	forceUserAcceptanceArgumentInMetricsClient = true

	shouldFlushMetricsClientQueueOnEachEvent = true
)

func initInteractiveConfig() (*KurtosisConfig, error) {

	fmt.Println(metrics_optin.WhyKurtosisCollectMetricsDescriptionNote)

	didUserAcceptSendingThatThemRejectSendingMetrics := false

	didUserAcceptSendingMetrics, err := prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsPromptLabel, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
	}

	if !didUserAcceptSendingMetrics {
		didUserAcceptSendingThatThemRejectSendingMetrics, err = prompt_displayer.DisplayConfirmationPromptAndGetBooleanResult(metricsConsentPromptLabel, true)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred displaying user metrics consent prompt")
		}
	}

	if didUserAcceptSendingMetrics || didUserAcceptSendingThatThemRejectSendingMetrics{
		metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()

		metricsUserId, err := metricsUserIdStore.GetUserID()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting metrics user ID")
		}

		//We pass forceUserAcceptanceArgumentInMetricsClient argument here because if we don't do this, the metrics client will be "DoNothing"
		metricsClient, err := metrics_client.CreateMetricsClient(source.KurtosisCLISource, kurtosis_cli_version.KurtosisCLIVersion, metricsUserId, forceUserAcceptanceArgumentInMetricsClient, shouldFlushMetricsClientQueueOnEachEvent)
		if err != nil {
			return nil,  stacktrace.Propagate(err, "An error occurred creating the metrics client")
		}
		defer func() {
			if err := metricsClient.Close(); err != nil {
				logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
			}
		}()

		if err := metricsClient.TrackShouldSendMetricsUserElection(didUserAcceptSendingMetrics); err != nil {
			//We don't want to interrupt users flow if something fails when tracking metrics
			logrus.Errorf("An error occurred tracking should send metrics user election event\n%v",err)
		}
	}

	kurtosisConfig := NewKurtosisConfig(didUserAcceptSendingMetrics)
	return kurtosisConfig, nil
}

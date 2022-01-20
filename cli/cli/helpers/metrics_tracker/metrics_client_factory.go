package metrics_tracker

import (
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client/snow_plow_client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	"github.com/kurtosis-tech/stacktrace"
)

func CreateMetricsClient() (client.MetricsClient, error) {

	metricsUserIdStore := NewMetricsUserIDStore()

	userId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting metrics user id")
	}

	metricsClient, err := snow_plow_client.NewSnowPlowClient(source.KurtosisCLISource, userId)
	if err != nil {
		//We don't throw and error if this fails, because we don't want to interrupt user's execution
		return nil, stacktrace.Propagate(err, "An error occurred creating SnowPlow metrics client")
	}

	return metricsClient, nil
}

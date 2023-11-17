package metrics_cloud_user_instance_id_helper

import (
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/sirupsen/logrus"
)

func GetMaybeCloudUserAndInstanceID() (metrics_client.CloudUserID, metrics_client.CloudInstanceID) {
	cloudUserId := ""
	cloudInstanceId := ""
	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		logrus.Debugf("Could not retrieve the current context")
		logrus.Debugf("Error was: %v", err.Error())
	} else {
		if store.IsRemote(currentContext) {
			cloudUserId = currentContext.GetRemoteContextV0().GetCloudUserId()
			cloudInstanceId = currentContext.GetRemoteContextV0().GetCloudInstanceId()
		}
	}
	return metrics_client.CloudUserID(cloudUserId), metrics_client.CloudInstanceID(cloudInstanceId)
}

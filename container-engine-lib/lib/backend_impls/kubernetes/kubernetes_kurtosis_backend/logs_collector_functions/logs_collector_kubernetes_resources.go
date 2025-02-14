package logs_collector_functions

import (
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type logsCollectorKubernetesResources struct {
	daemonSet *v1.DaemonSet

	// store all pods associated with fluent bit log collectors?
	// could make it easier to do health and status checks?
	//pods *apiv1.Pod

	configMap *apiv1.ConfigMap
}

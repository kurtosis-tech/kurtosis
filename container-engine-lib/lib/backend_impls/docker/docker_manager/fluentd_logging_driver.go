package docker_manager

import (
	"github.com/docker/docker/api/types/container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components"
	"strings"
)

const (
	//We almost could have gotten these values from the Docker package github.com/docker/docker/daemon/logger/fluentd but, those are private so, we have to redeclare it
	fluentdLoggingDriverTypeName         = "fluentd"
	fluentdLoggingDriverAddressConfigKey = "fluentd-address"
	loggingDriverLabelsKey = "labels"

	labelsSeparator = ","
)

type fluentdLoggingDriver struct {
	address logs_components.LogsCollectorAddress
	labels logs_components.LogsCollectorLabels
}

func NewFluentdLoggingDriver(logsCollectorAddress logs_components.LogsCollectorAddress, labels logs_components.LogsCollectorLabels) *fluentdLoggingDriver {
	return &fluentdLoggingDriver{
		address: logsCollectorAddress,
		labels: labels,
	}
}

func (config *fluentdLoggingDriver) GetLogConfig() container.LogConfig {
	return container.LogConfig{
		Type: fluentdLoggingDriverTypeName,
		Config: map[string]string{
			fluentdLoggingDriverAddressConfigKey: string(config.address),
			loggingDriverLabelsKey: config.getLabelsStr(),
		},
	}
}

func (config *fluentdLoggingDriver) getLabelsStr() string {
	return strings.Join(config.labels, labelsSeparator)
}

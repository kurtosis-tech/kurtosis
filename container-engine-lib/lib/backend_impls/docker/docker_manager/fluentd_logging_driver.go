package docker_manager

import (
	"github.com/docker/docker/api/types/container"
	"strings"
)

const (
	//We almost could have gotten these values from the Docker package github.com/docker/docker/daemon/logger/fluentd but, those are private so, we have to redeclare it
	fluentdLoggingDriverTypeName         = "fluentd"
	fluentdLoggingDriverAddressConfigKey = "fluentd-address"
	fluentdLoggingDriverAsyncConfigKey   = "fluentd-async"
	loggingDriverLabelsKey               = "labels"

	//Using async for do not blocking the container's logs if the Fluentd/Fluentbit service is down
	enableAsyncFluentbitLoggingDriverByDefault = "true"

	labelsSeparator = ","
)

type fluentdLoggingDriver struct {
	address string
	async   string
	labels  []string
}

func NewFluentdLoggingDriver(address string, labels []string) *fluentdLoggingDriver {
	return &fluentdLoggingDriver{
		address: address,
		async:   enableAsyncFluentbitLoggingDriverByDefault,
		labels:  labels,
	}
}

func (config *fluentdLoggingDriver) GetLogConfig() container.LogConfig {
	return container.LogConfig{
		Type: fluentdLoggingDriverTypeName,
		Config: map[string]string{
			fluentdLoggingDriverAddressConfigKey: config.address,
			fluentdLoggingDriverAsyncConfigKey:   config.async,
			loggingDriverLabelsKey:               config.getLabelsStr(),
		},
	}
}

func (config *fluentdLoggingDriver) getLabelsStr() string {
	return strings.Join(config.labels, labelsSeparator)
}

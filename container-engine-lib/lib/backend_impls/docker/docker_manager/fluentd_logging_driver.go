package docker_manager

import (
	"github.com/docker/docker/api/types/container"
	"strings"
)

const (
	//We get these values front the Docker package github.com/docker/docker/daemon/logger/fluentd but, those are private so, we have to redeclare it
	fluentdLoggingDriverTypeName         = "fluentd"
	fluentdLoggingDriverAddressConfigKey = "fluentd-address"
	loggingDriverLabelsKey = "labels"

	labelsSeparator = ","
)

type fluentdLoggingDriver struct {
	address string
	labels []string
}

func NewFluentdLoggingDriver(address string, labels []string) *fluentdLoggingDriver {
	return &fluentdLoggingDriver{address: address, labels: labels}
}

func (config *fluentdLoggingDriver) GetLogConfig() container.LogConfig {
	return container.LogConfig{
		Type: fluentdLoggingDriverTypeName,
		Config: map[string]string{
			fluentdLoggingDriverAddressConfigKey: config.address,
			loggingDriverLabelsKey: config.getLabelsStr(),
		},
	}
}

func (config *fluentdLoggingDriver) getLabelsStr() string {
	return strings.Join(config.labels, labelsSeparator)
}

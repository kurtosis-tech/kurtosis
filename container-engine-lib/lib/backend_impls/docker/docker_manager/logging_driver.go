package docker_manager

import "github.com/docker/docker/api/types/container"

type LoggingDriver interface {
	GetLogConfig() container.LogConfig
}

package docker_manager

import "github.com/docker/docker/api/types/container"

type loggingDriver interface {
	GetLogConfig() container.LogConfig
}

package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"time"
)

type PerHourFileLayout struct {
	clock logs_clock.LogsClock
}

func NewPerHourFileLayout(clock logs_clock.LogsClock) *PerHourFileLayout {
	return &PerHourFileLayout{clock: clock}
}

func (phf *PerHourFileLayout) GetLogFileLayoutFormat() string {
	return ""
}

func (phf *PerHourFileLayout) GetLogFilePaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriod time.Duration,
	retentionPeriodIntervals int,
	enclaveUuid, serviceUuid string) []string {
	return []string{}
}

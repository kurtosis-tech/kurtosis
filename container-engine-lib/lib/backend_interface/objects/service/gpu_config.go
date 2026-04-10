package service

// GpuConfig bundles all GPU-related container configuration: which GPUs to expose,
// the shared-memory size, and ulimits (all of which are only meaningful for GPU workloads).
type GpuConfig struct {
	count            int64
	deviceIDs        []string
	shmSizeMegabytes uint64
	ulimits          map[string]int64
}

func NewGpuConfig(count int64, deviceIDs []string, shmSizeMegabytes uint64, ulimits map[string]int64) GpuConfig {
	return GpuConfig{
		count:            count,
		deviceIDs:        deviceIDs,
		shmSizeMegabytes: shmSizeMegabytes,
		ulimits:          ulimits,
	}
}

func (g GpuConfig) GetCount() int64 {
	return g.count
}

func (g GpuConfig) GetDeviceIDs() []string {
	return g.deviceIDs
}

func (g GpuConfig) GetShmSizeMegabytes() uint64 {
	return g.shmSizeMegabytes
}

func (g GpuConfig) GetUlimits() map[string]int64 {
	return g.ulimits
}

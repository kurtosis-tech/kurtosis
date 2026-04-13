package service

const (
	// DefaultDockerGpuDriver is the Docker DeviceRequest driver used when none is specified.
	DefaultDockerGpuDriver = "nvidia"
	// DefaultK8sGpuResourceName is the Kubernetes resource name used when none is specified.
	DefaultK8sGpuResourceName = "nvidia.com/gpu"
)

// GpuConfig bundles all GPU-related container configuration: which GPUs to expose,
// the shared-memory size, and ulimits (all of which are only meaningful for GPU workloads).
type GpuConfig struct {
	count            int64
	deviceIDs        []string
	shmSizeMegabytes uint64
	ulimits          map[string]int64
	dockerDriver     string
	k8sResourceName  string
}

func NewGpuConfig(count int64, deviceIDs []string, shmSizeMegabytes uint64, ulimits map[string]int64, dockerDriver string, k8sResourceName string) GpuConfig {
	return GpuConfig{
		count:            count,
		deviceIDs:        deviceIDs,
		shmSizeMegabytes: shmSizeMegabytes,
		ulimits:          ulimits,
		dockerDriver:     dockerDriver,
		k8sResourceName:  k8sResourceName,
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

// GetDockerDriver returns the Docker DeviceRequest driver name (e.g. "nvidia", "amd").
func (g GpuConfig) GetDockerDriver() string {
	if g.dockerDriver == "" {
		return DefaultDockerGpuDriver
	}
	return g.dockerDriver
}

// GetK8sResourceName returns the Kubernetes resource name for GPU requests (e.g. "nvidia.com/gpu").
func (g GpuConfig) GetK8sResourceName() string {
	if g.k8sResourceName == "" {
		return DefaultK8sGpuResourceName
	}
	return g.k8sResourceName
}

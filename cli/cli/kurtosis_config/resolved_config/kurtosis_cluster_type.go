package resolved_config

//go:generate go run github.com/dmarkham/enumer -type=KurtosisClusterType -trimprefix=KurtosisClusterType_ -transform=snake
type KurtosisClusterType int
const (
	KurtosisClusterType_Docker KurtosisClusterType = iota
	KurtosisClusterType_Kubernetes
)

package commons

type JsonRpcServiceNetwork struct {
	NetworkId string
	ServiceContainerIds map[int]string
	ServiceIps map[int]string
	// TODO might be better to make this a nat.Port object
	ServiceJsonRpcPorts map[int]int
	// TODO might be better to make this a nat.Port object
	ServiceCustomPorts map[int]map[int]int

	// The final liveness requests that need to be monitored before the network can be considered alive
	// In practice, these are the liveness requests of the leaves of the dependency graph
	NetworkLivenessRequests map[JsonRpcServiceSocket]JsonRpcRequest
}

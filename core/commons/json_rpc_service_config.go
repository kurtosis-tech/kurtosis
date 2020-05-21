package commons

// TODO probably split all this out into separate files

type JsonRpcServiceSocket struct {
	IPAddress string
	Port int
}

type JsonRpcVersion string
const (
	RPC_VERSION_1_0 = "1.0"
)

type JsonRpcRequest struct {
	Endpoint string
	Method string
	RpcVersion JsonRpcVersion
	Params map[string]string
	ID int
}
type ServiceSpecificPort int

type JsonRpcServiceConfig interface {
	GetDockerImage() string
	GetJsonRpcPort() int

	// Really this should be of type Map<? extends Enum, int> but Go doesn't have enums or generics :(
	// Thus, we have to rely on a user to invent their own "enum" and use that
	GetOtherPorts() map[ServiceSpecificPort]int

	// Should return a command to be run in the Docker container running the RPC service, with an image-appropriate
	// busy loop to wait for dependencies to come up
	GetContainerStartCommand() []string

	// Returns an object containing information about how to query this JSON rpc service for liveness
	GetLivenessRequest() *JsonRpcRequest
}
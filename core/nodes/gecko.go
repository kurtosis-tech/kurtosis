/*

Contains types to represent nodes contained in Docker containers.

*/

package nodes

// TODO make this an implementation of JsonRpcServiceConfig
// Type representing a Gecko Node and which ports on the host machine it will use for HTTP and Staking.
type GeckoNode struct {
	GeckoImageName, HttpPortOnHost, StakingPortOnHost string
}

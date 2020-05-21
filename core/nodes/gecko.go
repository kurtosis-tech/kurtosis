/*

Contains types to represent nodes contained in Docker containers.

*/

package nodes


// Type representing a Gecko Node and which ports on the host machine it will use for HTTP and Staking.
type GeckoNode struct {
	GeckoImageName, HttpPortOnHost, StakingPortOnHost string
	respID string
}

func (node *GeckoNode) GetRespID() string {
	return node.respID
}

// Creates a Docker container for the Gecko Node.
func (node *GeckoNode) Create() {

}

// Waits on the Docker container, and if it exits, exposes logs to stdout.

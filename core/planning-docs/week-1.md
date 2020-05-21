* [ ] BasicTestSuiteRunner that runs a basic image




Kurtosis Common API
-------------------
* [x] POGO class ServiceContainer with Constructor(String containerId)
* [x] POGO class JsonRpcLivenessRequest to represent the params needed to make a request to a service to check its liveness, containing:
    * Endpoint to query
    * Method
    * JSON RPC version to use
    * Params
* [x] Abstract class `JsonRpcServiceConfig<P extends Enum>`, representing the Docker config specific to a container that will a) wait for an optional dependency JSON RPC service to start and b) run a service that will start serving JSON RPC, listening on a set of ports. P = an enum of the ports used by the service outside HTTP
    * [x] Abstract method GetDockerImage() -> String
    * [x] Abstract method GetHttpPort() -> int to inform the user of this class what port the service config will be listening on inside the container
    * [x] Abstract method GetOtherPorts() -> `Map<P, int>` to inform the user of this class what other ports the service will be listening on inside the container
    * [x] Abstract method GetStartCommand(`Map<Pair<String, int>, JsonRpcLivenessRequest>` dependencies) that should return the command line call needed to start this Docker image, with dependencies possibly empty if this node has no dependencies
        * NOTE: if necessary and the command has dependencies, this command should use the information about dependencies to include an image-specific busy loop for checking if the boot nodes are up first
    * [x] Abstract method GetLivenessRequest() -> JsonRpcLivenessRequest to inform users of this class what request they should make to the HTTP port to verify liveness
* [NOT THIS WEEK] Class HostPortUsageTracker, a very simple and constantly-running service that keeps track of what ports are already used on the host with:
    * Constructor(int acceptablePortStart, int acceptablePortEnd) that stores the params initializes a new `Set<int>` of usedPorts
    * Method GetNextFreePort() -> int, which maybe checks that the port it thinks is free is actually free before returning it?
* [ ] Class JsonRpcServiceNetwork, a POGO that stores:
    * `Map<int, String>` which maps serviceId -> IP addresses inside Docker's network
    * `Map<int, int>` mapping serviceId -> http port
    * `Map<int, Map<? extends Enum, int>>` mapping serviceId -> custom ports
    * `Map<Pair<String, int>, JsonRpcLivenessRequest>` containing the final liveness requests that need to be monitored for the network to be guaranteed alive (which are all the leaves of the config's DAG)
    * ...and getters....
* [ ] Class JsonRpcServiceNetworkConfig with:
    * [ ] Builder with:
        * [ ] Constructor(String networkId)
        * [ ] Some variation of AddService(int serviceId, JsonRpcServiceConfig config, `Set<int>` dependencies), that:
            1. Validates that ID hasn't been used yet
            2. Validates no circular dependencies????
            3. Stores the JsonRpcServiceConfig in a `Map<int, JsonRpcServiceConfig>` nodes
            4. Stores the dependencies in a DAG
        * NOTE: we'll create an AddLink call later, that will allow us to modify the network configuration of nodes
    * [ ] Method Initialize(DockerSdk docker, HostPortUsageTracker tracker) -> `Map<int, Pair<String, int>>` which:
        1. Create new Docker network with networkID
        2. TODO make any specific networking customizations (e.g. latency between nodes)
        3. Start root nodes of DAG, using the port tracker to map host ports to ports the services are listening on in their containers
        4. Start next level, who should already have wait commands built into their startup time, passing in root nodes as dependencies
        5. ....etc...
        6. Returns mapping of serviceId -> (ipAddr on Docker network, port)
* [ ] Interface TestNetworkConfigProvider with:
    * [ ] Method GetConfig() -> JsonRpcServiceNetworkConfig that returns the config used to create the Docker containers
    * [ ] Method GetType() -> `<E extends Enum>`

Ava Impl of Commons API
-----------------------
### Ava Service Implementations
* [ ] Enum GeckoV1Ports:
    * STAKING_PORT
* [ ] Class GeckoV1ServiceConfig implements JsonRpcServiceConfig<GeckoV1Ports>:
    * [ ] Constructor(String dockerImage, ...Gecko v1-particular params...)
    * [ ] Method GetJsonRpcPort() returns 9650
    * [ ] Method GetOtherPorts() returns a map of STAKING_PORT -> 9651
    * [ ] Method GetStartCommand that uses the constructor params to return a command, runnable by the Docker container that Gecko is based on, to check the liveness of the dependencies using the given liveness requests (likely a Bash while loop)
    * [ ] Method GetLivenessRequest() returns a JsonRpcLivenessRequest representing the getCurrentValidators API call

### Ava Network Implementations
TODO Doesn't need to be exactly like this
* [ ] Class TenNodeNetworkNodeIdx
    * ONE
    * TWO
    * ....
    * TEN
* [ ] Class TenNodeGeckoTestNetwork implements TestNetwork, with:
    * [ ] Constructor(JsonRpcServiceNetwork network)
    * [ ] GetNode(int x) -> AvaV1GoBinding to allow a test to make calls to a node with the given index
    * [ ] AssertAllNodesHaveStateX(..some state..) -> boolean
* [ ] Enum AvaNetworkType implements TestNetworkConfigProvider, TestNetworkDeserializer:
    * [ ] HOMOGENOUS_2_NODE_GECKO_V1
        * [ ] GetNetworkConfig() to initialize a network of 2 x v1 Gecko nodes, doing the boot node first
        * [ ] GetType() returns this
        * [ ] GetTestNetwork that creates a new Homogenous2NodeGeckoV1TestNetwork
    * HETEROGENOUS_2_NODE_GECKO
        * [ ] GetNetworkConfig() to initialize a network of 2 nodes, one on V1 and one on V2
        * [ ] GetType() returns this
        * [ ] GetTestNetwork that creates a new Heterogenous2NodeGeckoTestNetwork
* [ ] AvaTestNetworkDeserializer implements TestNetworkDeserializer<AvaNetworkType> with:
    * [ ] Method GetTestNetworkClass that returns the appropriate TestNetwork type, either two-node or ten-node, based off the network type
    * [ ] Method GetTestNetwork that instantiates the appropriate class based off the enum input

Kurtosis Test Initializer Library
---------------------------------
The idea here is that the clients wrap this initializer in a very thin library of their own that process command-line args however they please.

* [ ] Class TestSuiteRunner that that provides a class, meant to be embedded in a CLI, for running the tests
    * [ ] Constructor(String testControllerDockerImageRef, `Map<String, TestNetworkConfigProvider>` tests)
    * [ ] Method RunTests(DockerSdk docker, List<String> testIds, int parallelism) that:
        1. [ ] Retrieves the given tests from the map
        1. [ ] Spins up threads = parallelism, each of which:
            1. Pulls the corresponding TestNetworkConfigProvider for the test being run
            2. Runs TestNetworkConfigProvider.GetNetworkConfig().Initialize(docker) to create the cluster and get back a JsonRpcNetwork object with its dimensions
            3. Serializes the JsonRpcNetwork defining the test network's state to YAML
            4. Creates a test controller Docker VM with the YAML file mounted and the enum representing the type of network passed to the controller so it can construct the appropriate TestNetwork

Ava Impl of Test Initializer Library
------------------------------------
This library depends on the Ava implementation of the Kurtosis common API!

* [ ] Class AvaE2ETestInitializer that:
    1. Creates a TestSuiteRunner with a mapping of all the test names -> TwoNodeGeckoV1TestNetwork or TenNodeGeckoTestNetwork as appropriate
    1. Reads CLI flags to figure out which tests are desired
    1. Call TestSuiteRunner.RunTests()



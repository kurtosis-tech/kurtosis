Overview
========
Components
----------
1. Test controller, which executes a given test
2. Cluster Docker containers, which run the Ava nodes
3. The initializer, which is responsible for:
    * Serving as the entrypoint, called via CI
    * Spinning up the Ava network cluster
    * Spinning up the test controller, passing the necessary information across to it
    * Retrieving information after the test controller is done and reporting it to CI

Dependency Flowchart
--------------------
```
   COMMONS API
       ^
       |
   depends on
       |
  TEST CONTROLLER
       ^
       |
   depends on
       |
 TEST INITIALIZER
```


Outcomes Needed
===============
Ava Client Library
------------------
TODO; golang client bindings go here

Kurtosis Common API
-------------------
This is a library that is included by both the test initializer and the test controller, and serves as the bridge between the two. It's mostly responsible for structuring information about the running Ava Docker container, as that's the shared bit of information that the two talk over.

This might be able to be replaced by either Docker Compose templating, or Golang docker-compose bindings... but [the only ones I've found are deprecated](https://github.com/callicoder/go-docker-compose).

* [ ] POGO class ServiceContainer with Constructor(String containerId)
* [ ] POGO class LivenessRequest to represent the params needed to make a request to a service to check its liveness, containing:
    * Endpoint to query
    * Method
    * JSON RPC version to use
    * Params
* [ ] Abstract class `JsonRpcServiceConfig<P extends Enum>`, representing the Docker config specific to a container that will a) wait for an optional dependency JSON RPC service to start and b) run a service that will start serving JSON RPC, listening on a set of ports. P = an enum of the ports used by the service outside HTTP
    * [ ] Abstract method GetDockerImage() -> String
    * [ ] Abstract method GetHttpPort() -> int to inform the user of this class what port the service config will be listening on inside the container
    * [ ] Abstract method GetOtherPorts() -> `Map<P, int>` to inform the user of this class what other ports the service will be listening on inside the container
    * [ ] Abstract method GetStartCommand(`Map<Pair<String, int>`, LivenessRequest> dependencies) that should return the command line call needed to start this Docker image, with dependencies possibly empty if this node has no dependencies
        * NOTE: if necessary and the command has dependencies, this command should use the information about dependencies to include an image-specific busy loop for checking if the boot nodes are up first
    * [ ] Abstract method GetLivenessRequest() -> LivenessRequest to inform users of this class what request they should make to the HTTP port to verify liveness
    * [ ] Method Initialize(Docker docker) -> ServiceContainer which instantiates a service using the methods on this config
* [ ] Class HostPortUsageTracker, a very simple and constantly-running service that keeps track of what ports are already used on the host with:
    * Constructor(int acceptablePortStart, int acceptablePortEnd) that stores the params initializes a new `Set<int>` of usedPorts
    * Method GetNextFreePort() -> int, which maybe checks that the port it thinks is free is actually free before returning it?
* [ ] Class JsonRpcServiceNetwork, a POGO that stores:
    * `Map<int, String>` which maps serviceId -> IP addresses inside Docker's network
    * `Map<int, int>` mapping serviceId -> http port
    * `Map<int, Map<? extends Enum, int>>` mapping serviceId -> custom ports
    * `Map<Pair<String, int>, LivenessRequest>` containing the final liveness requests that need to be monitored for the network to be guaranteed alive (which are all the leaves of the config's DAG)
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

NOTE: THE BELOW IS UGLY AND NEEDS TO BE CHANGED! The basic idea is, it's the bridge between the initializer saying "this is the type of Docker container cluster I initialized" and the test controller going "gotcha, I recognize this type of cluster so I know what class to wrap this in". I'm guessing the initializer will pass this information to the test controller by serializing the JsonRpcServiceNetwork to file and mounting it on the test controller's container.

* [ ] Interface TestNetworkConfigProvider with:
    * [ ] Method GetConfig() -> JsonRpcServiceNetworkConfig that returns the config used to create the Docker containers
    * [ ] Method GetType() -> `<E extends Enum>`
* [ ] Interface `TestNetworkDeserializer<E extends Enum>`, which is a class that will be used on the test controller side for reading the information that the initializer gave it about what type of Docker cluster exists, with:
    * [ ] Method GetTestNetworkClass(E extends Enum) -> `Class<? extends TestNetwork>` , so that the user knows which specific type to cast the TestNetwork to after it comes out the other side
    * [ ] Method GetTestNetwork(E extends Enum, JsonRpcServiceNetwork network) -> TestNetwork, which uses GetTestNetworkClass to determine the right class and instantiate the right user impl of TestNetwork

// Enum that:
    // Gets the config for a given test network
    // Takes a running network and transforms it into a user-custom implementation of the class

Ava Impl of Commons API
-----------------------
### Ava Service Implementations
* [ ] Enum GeckoV1Ports:
    * STAKING_PORT
* [ ] Class GeckoV1ServiceConfig implements JsonRpcServiceConfig<GeckoV1Ports>:
    * [ ] Constructor(String dockerImage, ...Gecko v1-particular params...)
    * [ ] Method GetHttpPort() returns 9650
    * [ ] Method GetOtherPorts() returns a map of STAKING_PORT -> 9651
    * [ ] Method GetStartCommand that uses the constructor params to return a command, runnable by the Docker container that Gecko is based on, to check the liveness of the dependencies using the given liveness requests (likely a Bash while loop)
    * [ ] Method GetLivenessRequest() returns a LivenessRequest representing the getCurrentValidators API call

### Ava Network Implementations
* [ ] Class TwoNodeGeckoV1TestNetwork implements TestNetwork, with:
    * [ ] Constructor(JsonRpcServiceNetwork network)
    * [ ] GetBootNode() -> AvaV1GoBinding to allow a test to make calls to the first node
    * [ ] GetSecondNode() -> AvaV1GoBinding to allow a test to make calls to the second node
    * [ ] AssertBothNodesHaveStateX(..some state..) -> boolean
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

Kurtosis Test Controller API
-------------------
Depends on the Kurtosis Commons API

* [ ] Interface `Test<T extends TestNetwork>`, with:
    * [ ] Method Run(T network) -> boolean (?) where the user implements their specific logic to run tets
* [ ] Abstract class `TestControllerCli<E extends Enum>` with:
    * [ ] Constructor TestControllerCli(`Map<String, Test>` registeredTests, `TestNetworkDeserializer<E>` deserializer)
    * [ ] Abstract method GetDeserializer() -> TestNetworkDeserializer
    * [ ] Method Main(String[] args) -> void func with:
        * Args:
            1. A file, mounted on the test Docker, that contains YAML representing the JsonRpcServiceNetwork that the Docker environment was created with
            1. An enum in String format representing the type of Docker test network cluster that's been created
            1. The name of the test to run
        * Logic:
            1. Deserialize the YAML into a JsonRpcServiceNetwork
            1. Deserialize the Enum into E
            2. Call the TestNetworkDeserializer.GetTestNetwork(E, JsonRpcServiceNetwork) to get the right TestNetwork from what was intantiated
                * NOTE: this is ugly, and I know there's a better way to do it but I'm not fussing with it now
            3. Fetch the Test object from the map based on what the user requested
            4. Call Test.Run using the network type that we casted

Ava Impl of Test Controller
---------------------------
Depends on the Ava impl of the Kurtosis commons API

* [ ] Class TwoNodeNetworkTest<TwoNodeGeckoV1TestNetwork>:
    * ...does some shit against the two-node network...
* [ ] Class TenNodeNetworkTest<TenNodeGeckoV1TestNetwork>:
    * ...does some shit against the ten-node network...
* [ ] Class AvaTestControllerCli whose constructor takes in the mapping of tests and the AvaTestNetworkDeserializer

Kurtosis Test Initializer Library
---------------------------------
The idea here is that the clients wrap this initializer in a very thin library of their own that process command-line args however they please.

* [ ] Class TestSuiteRunner that that provides a class, meant to be embedded in a CLI, for running the tests
    * [ ] Constructor(String testControllerDockerImageRef, `Map<String, TestNetworkConfigProvider>` tests)
    * [ ] Method RunTests(DockerSdk docker, List<String> testIds, int parallelism) that:
        1. Retrieves the given tests from the map
        1. Spins up threads = parallelism, each of which:
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



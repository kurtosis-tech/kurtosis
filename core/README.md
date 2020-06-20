# Kurtosis
Kurtosis is a library framework for end-to-end testing distributed systems in Docker.

## Architecture
The Kurtosis architecture has three components:

1. An **initializer**, which is responsible for spinning up the node networks required by the tests
2. The **network**, composed of nodes running in Docker with user-defined parameters
3. The **controller**, a Docker image running the code that actually makes requests to the network to run the tests

The control flow goes:

1. The initializer launches and looks at what tests need to be run
2. For each test:
    1. The initializer spins up the Docker images of the network that the test requires
    2. The initializer spins up a Docker image of the controller, passing to the controller the test name and network IP information
    3. The controller waits for the network to become available (all nodes respond to their user-defined liveness requests)
    4. The controller runs the desired test while the initializer waits for it to complete
    5. The controller returns the result and exits
    6. The initializer tears down the test network
4. After all tests are run, the initializer returns the results of the tests

## Getting Started
At a high level, you'll need the following to start writing tests:

1. One or more network definitions, to tell the initializer what shape of network to spin up
2. One or more tests that consume those networks, make calls to the networks, and fail if necessary
3. A `main.go` file that runs the controller code
4. A `main.go` file that runs the initializer

More concretely, you'll need at least:

### Commons
1. A struct representing a node in your network (identified by an IP address), with functions for the calls that can be made to the node
    * NOTE: Tests will use this object, so it should have whatever methods needed to make test-writing clean and easy
1. An implementation of the `ServiceInitializerCore` interface to handle the particulars of spinning up the Docker image running your service
1. An implementation of the `ServiceAvailabilityCheckerCore` interface to handle querying your service to see if it's available yet
1. A struct representing a specific instantion of a network (e.g. `ThreeNodeNetwork`), composed of the nodes you've defined
    * NOTE This object will be passed directly to the tests, so it should be given whichever methods make test-writing clean and easy (e.g. `GetNodeOne()`, `GetNodeTwo()`, etc.)
1. A struct implementing the `TestNetworkLoader` interface for bootstrapping your network, which defines:
    1. How to create the network out of Docker images (using initializer & availability checker cores you defined)
    2. How to construct the network out of a set of IPs
1. A test implementing the `Test` class
    * NOTE: Because Go doesn't have generics yet, the first line of your `Run` method should be casting the generic network to your custom-defined struct representing the type of network that the test consumes
1. A struct implementing the `TestSuite` class, which declares the tests that are available

### Initializer
1. A `main.go` for the initializer that constructs an instance of `TestSuiteRunner` (which handles all the initialization)

### Controller
1. A `main.go` for the controller that constructs an instance of `TestController` for running a given test
1. A Docker image that runs the controller's `main.go`, which will be launched by the initializer
    * NOTE: the initializer will set two special environment variables in the controller image's shell: `TEST_NAME` and `NETWORK_DATA_FILEPATH` (TODO list out all of them!). These should be consumed by your controller's `main.go` and passed as-is to the `NewTestController` call.

Some implementation tips:
* We recommend structuring your code into the same `commons`, `initializer`, and `controller` packages listed above.
* Make an interface representing the calls that every single node on your network can receive (e.g. `ElasticsearchService`), and create sub-interfaces for more specific calls (e.g. `ElasticsearchMasterService`)
* Both network-creating and network-loading functions are intentionally centralized in the `TestNetworkLoader` interface so that constant IDs can be defined for each node and used in both network creation & loading

## Examples
See [the Ava end-to-end tests](https://github.com/kurtosis-tech/ava-e2e-tests) for the reference Kurtosis implementation

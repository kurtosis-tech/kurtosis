# Kurtosis
Kurtosis is a framework for writing end-to-end test suites for distributed systems using Docker.

## Architecture
The Kurtosis architecture has four components:

1. The **test network**, composed of Docker containers running the services necessary for a given test
1. The **test suite**, the package of tests that can be run
1. The **initializer**, which is the entrypoint to the application and responsible for running the tests in the test suite
1. The **controller**, the Docker container responsible for orchestrating the execution of a single test (including spinning up the test network)

The control flow goes:

1. The initializer launches and looks at what tests need to be run
1. For each test, with the desired amount of parallelism:
    1. The initializer launches a controller Docker container
    1. The controller spins up a network of whichever Docker images the test requires
    1. The controller waits for the network to become available
    1. The controller runs the desired test
    1. After the test finishes, the controller tears down the network of test-specific containers
    1. The controller returns the result to the initializer and exits
1. The initializer waits for all tests to complete and returns the results

## Getting Started
To run tests with Kurtosis, you'll need to define custom components for producing each of the four Kurtosis components. At a high level, this means writing:

1. A test network definition that declares the set of Docker images that a test will use
1. A test suite package of tests for your application
1. A CLI that calls down to Kurtosis' initializer code
1. A Docker image that runs a CLI that calls down to Kurtosis' controller code

More specifically, for each of the components you'll need:

### Test Network Definition
* For each of the services in your test networks, you'll need interfaces representing the interactions possible on the services in your network, e.g.:
    ```
    type MyService interface {
        GetRpcPort() 



1. One or more implementations of the 
At a high level, you'll need the following to start writing tests:

1. One or more network definitions, to tell the controller what shape of service network you'll need for your tests
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
    * NOTE: the initializer will set several environment variables in the controller image's shell environment that the controller Docker image should use. For the most up-to-date information about what those environment variables are, see the `generateTestControllerEnvVariables` function [here](https://github.com/kurtosis-tech/kurtosis/blob/develop/initializer/parallelism/test_executor.go).

Some implementation tips:
* We recommend structuring your code into the same `commons`, `initializer`, and `controller` packages listed above.
* Make an interface representing the calls that every single node on your network can receive (e.g. `ElasticsearchService`), and create sub-interfaces for more specific calls (e.g. `ElasticsearchMasterService`)
* Both network-creating and network-loading functions are intentionally centralized in the `TestNetworkLoader` interface so that constant IDs can be defined for each node and used in both network creation & loading

### Notes
While running, Kurtosis will create the following, per test:
* A new Docker network for the test
* A new Docker volume to pass files relevant to the test in
* Several containers related to the test

**If Kurtosis is killed abnormally (e.g. SIGKILL or SIGQUIT), the user will need to remove the Docker network and stop the running containers!** The specifics will depend on what Docker containers you start, but can be done using something like the following examples:

Find & remove Kurtosis Docker networks:
```
docker network ls  # See which Docker networks are left around - will be in the format of UUID-TESTNAME
docker network rm some_id_1 some_id_2 ...
```

**If the network isn't removed, you'll get IP conflict errors from Docker on next Kurtosis run!**

Stop running containers:
```
docker container ls    # See which Docker containers are left around - these will depend on the containers spun up
docker stop $(docker ps -a --quiet --filter ancestor="IMAGENAME" --format="{{.ID}}")
```


If Kurtosis is allowed to finish normally, the Docker network will be deleted and the containers stopped. **However, even with normal exit, Kurtosis will not delete the Docker containers or volume it created.** This is intentional, so that a dev writing Kurtosis tests can examine the containers and volume that Kurtosis spins up for additional information. It is therefore recommended that the user periodically clear out their old containers, volumes, and images; this can be done with something like the following examples:

Stopping & removing containers:
```
docker rm $(docker stop $(docker ps -a -q --filter ancestor="IMAGENAME" --format="{{.ID}}"))
```

Remove all volumes associated with a given test:
```
docker volume rm $(docker volume ls | grep "TESTNAME" | awk '{print $1}')
```

Remove unused images:
```
docker image rm $(docker images --quiet --filter "dangling=true")
```

## Examples
See [the Ava end-to-end tests](https://github.com/kurtosis-tech/ava-e2e-tests) for the reference Kurtosis implementation

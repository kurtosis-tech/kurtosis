# Kurtosis Implementation Tutorial
To run tests with Kurtosis, we'll need to define custom components for producing each of the four Kurtosis components. At a high level, this means writing:

1. At least one **network** definition that our tests can use to declare the type of network they want to run against
1. A **test suite** with at least one test inside
1. A **controller** Docker image that runs code wrapping Kurtosis' controller library
1. An **initializer** CLI that wraps Kurtosis' initializer library

In the below tutorial we'll show how to implement each of these components from scratch!

## The Test Network
A test always runs against a network of services being tested, and each test will declare the type of network it wants to run against. We're  and we'll need to tell Kurtosis what that network should look like for the test we'll write in this tutorial. 

Networks are composed of services, so our first step is defining what a "service" looks like. 

In this example, we'll suppose that our network is composed of REST microservices running in Docker images, all communicating with each other. We'll start by defining an interface that represents the functions a test can call on a service node, which is accomplished by implementing the [Service](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service.go) marker interface like so:

```go
type ServiceSocket struct {
    IpAddr string
    Port nat.Port
}

type MyService interface {
    services.Service

    GetHttpRestSocket() ServiceSocket
}
```

Now we have an interface that a test could use to interact with services in our network, but we still need to provide Kurtosis with an implementation. Every service in Kurtosis is a Docker container, so our implementation needs to represent the actual service running in the Docker container. Fulfilling the `MyService` contract means returning the socket of the service, so our implementation needs to return the IP address of the Docker container and the port our service is listening on inside the container:

```go
type MyServiceImpl struct {
    IpAddr string
    Port nat.Port
}

func (s MyServiceImpl) GetHttpRestSocket() ServiceSocket {
    return ServiceSocket{IpAddr: s.IpAddr, Port: s.Port}
}
```

Our tests now have a nice interface for interacting with a service, and we cleanly represent the reality that the service is backed by a Docker container with an IP address, listening on a port.

Next, we need to give Kurtosis the details on how to actually start a Docker container running one of our services. Kurtosis makes this easy - we only need to fill out the [ServiceInitializerCore](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service_initializer_core.go) interface. This interface is well-documented, so we'll use the documentation to write a service initializer core for our service like so:

```go
const (
    // We'll imagine the Docker image running our service is hardcoded to listen on port 80
    httpRestPort nat.Port = "80/tcp"

    configFileKey = "config-file"
)

type MyServiceInitializerCore struct {
    // This is a parameter specific to our service
    MyServiceLogLevel string

    // We could put more service-specific parameters here if needed
}

func (core MyServiceInitializerCore) GetUsedPorts() map[nat.Port]bool {
    return map[nat.Port]bool{httpRestPort: true}
}

func (core MyServiceInitializerCore) GetServiceFromIp(ipAddr string) Service {
    return MyServiceImpl{IpAddr: ipAddr, Port: httpRestPort}
}

func (core MyServiceInitializerCore) GetFilesToMount() map[string]bool {
    return map[string]bool{
        configFileKey: true,
    }
}

func (core MyServiceInitializerCore) InitializeMountedFiles(mountedFiles map[string]*os.File, dependencies []Service) error {
    configFp := mountedFiles[configFileKey]
    // Do some initialization of our config file in preparation for service launch
    return nil
}

func (core MyServiceInitializerCore) GetTestVolumeMountpoint() string {
    // Return a filepath on our Docker image that's safe to mount the Kurtosis test volume on
}

func (core MyServiceInitializerCore) GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr string, dependencies []Service) ([]string, error) {
    // Use our specific knowledge of the Docker image to craft the command the Docker image will run with
    // This is where we'll use any service-specific params from above, including MyServiceLogLevel
}
```

Note the service "dependencies" that show up above. Kurtosis knows that some services will depend on others, and gives the developer the option to modify a service's files and start command based on other existing services in the network. We'll see how to declare these dependencies later.

Since a Docker container being up doesn't mean that the service inside is available and since we don't want to run a test against a network of services that are still starting up, the last piece we need for our service is a way to tell Kurtosis when the service is actually available for use. We'll therefore implement the [ServiceAvailabilityCheckerCore](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service_availability_checker_core.go) interface like so:

```go
type MyServiceAvailabilityCheckerCore struct {}

func (core MyServiceAvailabilityCheckerCore) IsServiceUp(toCheck Service, dependencies []Service) bool {
    // Go doesn't have generics, so we need to cast to our service interface
    castedToCheck := toCheck.(MyService)

    httpRestSocket := castedToCheck.GetHttpRestSocket()
    
    var isAvailable bool

    // ...Use the socket to query some healthcheck endpoint to verify our service is up and set isAvailable accordingly

    return isAvailable
}

func (core MyServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
    // This could also be parameterizable
    return 30 * time.Second
}
```

We're all set up to use this service in a network now... but of course, we still need to define what that network looks like. Kurtosis has a [ServiceNetwork](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/networks/service_network.go) object that represents the underlying state of the test network, but interacting with it is often too low-level for writing clean tests. To make writing tests as simple as possible, Kurtosis lets the developer define an arbitrary network struct that wraps the low-level network representation; this struct will then be passed to the tests. The developer can define this higher-level network wrapper object any way they please, but in our example we'll imagine that our tests all use a three-node network. Thus, our wrapper struct looks like so:

```go
type ThreeNodeNetwork struct {
    networks.Network

    BootNode MyService
    DependentNode1 MyService
    DependentNode2 MyService
}
```

Much like `MyService`, this object provides a test-friendly API for interacting with the network. But, also like `MyService`, we need to tell Kurtosis how to actually initialize this object. Kurtosis has an interface, [TestNetworkLoader](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_network_loader.go), for bridging the gap between Kurtosis' low-level `ServiceNetwork` and our custom wrapper, so we'll want to define a `TestNetworkLoader` implementation.

Before we write our implementation though, it's worth understanding how Kurtosis networks are configured. Each network has one or more **service configurations**, which serve as templates for the service instances that will comprise the network. These service configurations are defined by a configuration ID, a docker image, a service initializer core, and an availability checker, so if a network is composed of only one type of service then the network only needs one configuration; if a network is made up of many different types of services then it will need many configurations.

Using this information and the documentation on `TestNetworkLoader`, we can now write our `ThreeNodeNetworkLoader` implementation:

```go
const (
    configId = 0

    bootNodeServiceId = 0
    dependentNode1ServiceId = 1
    dependentNode2ServiceId = 2
)

type ThreeNodeNetworkLoader struct {
    DockerImage string
    MyServiceLogLevel string
}

func (loader ThreeNodeNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
    initializerCore := MyServiceInitializerCore{MyServiceLogLevel: loader.MyServiceLogLevel}
    checkerCore := MyServiceAvailabilityCheckerCore{}

    builder.AddConfiguration(configId, loader.DockerImage, initializerCore, checkerCore)
    return nil
}

func (loader ThreeNodeNetworkLoader) InitializeNetwork(network *ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
    result := map[int]services.ServiceAvailabilityChecker{}

    // Create the boot node using the configuration we defined earlier
    bootChecker, err := network.AddService(configId, bootNodeServiceId, map[int]bool{})
    // ... error-checking omitted ...
    result[bootNodeServiceId] = bootChecker

    // Define dependent nodes that depend on the boot node (Go doesn't have a set type, so a map[int]bool is used instead)
    // NOTE: Error-checking has been omitted
    dependentNode1Checker, err := network.AddService(configId, dependentNode1ServiceId, map[int]bool{bootNodeServiceId: true})
    result[dependentNode1ServiceId] = dependentNode1Checker
    dependentNode2Checker, err := network.AddService(configId, dependentNode2ServiceId, map[int]bool{bootNodeServiceId: true})
    result[dependentNode2ServiceId] = dependentNode2Checker

    return result, nil
}

func (loader ThreeNodeNetworkLoader) WrapNetwork(network *ServiceNetwork) (Network, error) {
    // By moving the low-level ServiceNetwork calls here, we remove the need for a test to know how to do this
    bootNodeService := network.GetService(bootNodeServiceId).Service.(MyService)
    dependentNode1Service := network.GetService(dependentNode1Service).Service.(MyService)
    dependentNode2Service := network.GetService(dependentNode2Service).Service.(MyService)

    return ThreeNodeNetwork{
        BootNode: bootNodeService,
        DependentNode1: dependentNode1Service,
        DependentNode2: dependentNode2Service,
    }
}
```

Here, we can see service dependencies being declared: we have a boot node that doesn't depend on other nodes (and so receives an empty dependency set), and two dependent nodes who depend on the boot node (and so declare a dependency set of the boot node service ID). 

The heavy lifting is finally done - we've declared a service with the appropriate initializer and availability checker cores, a network composed of that service, and a loader to wrap the low-level Kurtosis representation with a simpler, test-friendly version. Now we can write some tests!


## The Test Suite
A test suite is simply a package of tests, and a test is just a definition of the required test network and a chunk of logic that validates against it. To write a test we'll need to implement the [Test](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test.go) interface like so:

```go
type ThreeNodeNetworkTest1 struct {
    DockerImage string
    MyServiceLogLevel string
}

func (test ThreeNodeNetworkTest1) Run(network networks.Network, context TestContext) {
    // Because Go doesn't have generics, we unfortunately need to cast the network to our custom network as the first thing in every test
    castedNetwork := network.(ThreeNodeNetwork)

    bootService := castedNetwork.BootNode
    bootSocket := bootService.GetHttpRestSocket()

    var callSuccessful bool
    // Execute some call against the bootSocket to check the state
    context.AssertTrue(callSuccessful, errors.New("Expected a successful call to the boot node!"))
}

func (test ThreeNodeNetworkTest1) GetNetworkLoader() (networks.NetworkLoader, error) {
    return ThreeNodeNetworkLoader{DockerImage: test.DockerImage, MyServiceLogLevel: test.MyServiceLogLevel}
}

func (test ThreeNodeNetworkTest1) GetExecutionTimeout() time.Duration {
    return 30 * time.Second
}

func (test ThreeNodeNetworkTest1) GetSetupBuffer() time.Duration {
    return 60 * time.Second
}
```

Note that test failures are logged using the [TestContext](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_context.go) object, in a manner similar to Go's inbuilt `testing.T` object.

We have a test now, so we can implement the [TestSuite](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_suite.go) interface to package it:

```go
type MyTestSuite struct {
    DockerImage string
}

func (suite MyTestSuite) GetTests() map[string]Test {
    return map[string]Test {
        "threeNodeNetworkTest1": ThreeNodeNetworkTest1{
            DockerImage: suite.DockerImage,
            MyServiceLogLevel: "info",
        },
    }
}
```

Now that we have our test in our test suite, we need to give Kurtosis the tools it needs to run it!

## The Controller
To orchestrate all the steps required to run a single test, we need to provide Kurtosis the Docker image of a controller that will instantiate our test suite, create an instance of Kurtosis' [TestController](https://github.com/kurtosis-tech/kurtosis/blob/develop/controller/test_controller.go), and call the `RunTest` function to run our test. The controller must be a Docker image, so we'll need to write a main function that performs the needed work and a Dockerfile to generate an image to run our main function.

When we look at instantiating `TestController`, we notice that its constructor requires many arguments that we won't know how to provide. Fortunately, our controller container will be launched with these values via Docker environment variables. The complete list of the environment variables that our image will receive is defined in the `generateTestControllerEnvVariables` function inside [TestExecutor](https://github.com/kurtosis-tech/kurtosis/blob/develop/initializer/parallelism/test_executor.go), so we'll need to make sure we write a Dockerfile that uses them like so:

```
# ...image-specific Docker initialization things

# NOTE: Environment variables passed in as of 2020-07-19
CMD ./controller \
    --test=${TEST_NAME} \
    --subnet-mask=${SUBNET_MASK} \
    --docker-network=${NETWORK_NAME} \
    --gateway-ip=${GATEWAY_IP} \
    --log-level=${LOG_LEVEL} \
    --service-image-name=${SERVICE_IMAGE_NAME} \
    --test-controller-ip=${TEST_CONTROLLER_IP} \
    --test-volume=${TEST_VOLUME} \
    --test-volume-mountpoint=${TEST_VOLUME_MOUNTPOINT} &> ${LOG_FILEPATH}
```

Note that `SERVICE_IMAGE_NAME` is actually a custom variable that we defined! Kurtosis allows users to define custom Docker variables which will get passed to the controller so that custom information necessary to the test can be passed across; we'll see this variable get set later.

We'll then need to pipe these values to our main function to create our `TestController` and return an exit code appropriate to the test result:

```go
func main() {
    testNameArg := flag.String("test", "", "The name of the test the controller will run")
    subnetMaskArg := flag.String("subnet-mask", "", "The name of the subnet the controller will run in")
    // ... etc....

    testSuite := MyTestSuite{DockerImage: *serviceImageNameArg}
    controller := controller.NewTestController(
        *testVolumeArg,
        *testVolumeMountpointArg,
        *dockerNetworkArg,
        *subnetMaskArg,
        *gatewayIpArg,
        *testControllerIpArg,
        testSuite,
        *testNameArg)

    setupErr, testErr := controller.RunTest(*testNameArg)
    if setupErr != nil {
            fmt.Println(fmt.Sprintf("Test %v encountered an error during setup (test did not run):", *testNameArg))
            fmt.Println(setupErr)
            os.Exit(1)
    }
    if testErr != nil {
            fmt.Println(fmt.Sprintf("Test %v failed:", *testNameArg))
            fmt.Println(testErr)
            os.Exit(1)
    }
    fmt.Println(fmt.Sprintf("Test %v succeeded", *testNameArg))
}
```

Once we build the Docker image, we'll have a controller image the initializer can use to run our test. We still need to define the initializer itself though, which is our last step.

## The Initializer
Any E2E test suite should be runnable in CI, which means we need a concrete entrypoint that our CI system can call to run the suite. We'll therefore need to build a main function to actually run our suite. Kurtosis makes this very simple - just write a main function that creates an instance of our test suite, plug it into an instance of [TestSuiteRunner](https://github.com/kurtosis-tech/kurtosis/blob/develop/initializer/test_suite_runner.go) with the controller image that will run the test, and have the CLI return an exit code corresponding to test results:

```go
const (
    // Extra time to give each test for test network setup & teardown, *on top of* the per-test timeout
    additionalTestTimeoutBuffer = 60 * time.Second

    // Each test runs in its own Docker network, and the network will have capacity for 2 ^ networkWidthBits IP addresses, so this should be set high enough
    // so that no test runs out of IP addresses
    networkWidthBits = 8

    // The number of tests to run in parallel
    parallelism = 4
)

func main() {
    serviceImageNameArg := flag.String("serviceImage", "", "The Docker image of the services being tested")
    controllerImageNameArg := flag.String("controllerImage", "", "The Docker image of the controller that will run orchestrate the execution of a single test")
    testSuite := MyTestSuite{DockerImage: *serviceImageNameArg}
    testSuiteRunner := NewTestSuiteRunner(
        testSuite,
        *controllerImagNameArg,
        map[string]string{
            // Here we set the service image Docker environment variable that the controller consumes!
            "SERVICE_IMAGE_NAME": *serviceImageNameArg,
        },
        additionalTestTimeoutBuffer,
        networkWidthBits)

    // We specify an empty set of tests to run, so we'll run all of them
    allTestsSucceeded, error := testSuiteRunner.RunTests(map[string]bool{}, parallelism)
    if error != nil {
        logrus.Error("An error occurred running the tests:")
        logrus.Error(error)
        os.Exit(1)
    }

    if allTestsSucceeded {
        os.Exit(0)
    } else {
        os.Exit(1)
    }
}
```

Our test suite is now ready to go! Compiling and running our main function will:

1. Run the `ThreeNodeNetworkTest1` test, which will
1. Launch our controller image, which will
1. Initialize a network of three `MyService` nodes and 
1. Pass the `ThreeNodeNetwork` wrapper to our test where 
1. Our test-specific logic will get run

After the test returns or times out, the controller will:
1. Tear down the network and 
1. Return an exit code to our initializer, which will
1. Exit with the correct exit to signal the results to CI

We now have a basic E2E testing suite using Kurtosis!

## Final Notes
The example above is intended to give you a basic understanding of the moving parts in Kurtosis. Some features that Kurtosis includes which weren't covered here:
* Adding & removing services dynamically from the network during a test
* Passing custom values from the initializer CLI, to the test suite, down to the individual tests (e.g. if `MyServiceLogLevel` that had been parameterized in the test suite, rather than hardcoded to `info`)

For a more detailed example, take a look at [the first Kurtosis implementation for Ava labs](https://github.com/kurtosis-tech/ava-e2e-tests).

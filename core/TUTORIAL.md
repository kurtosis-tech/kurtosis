# Tutorial
To run tests with Kurtosis, you'll need to define custom components for producing each of the four Kurtosis components. At a high level, this means writing:

1. A test network definition that defines a network of services that a test will use
1. A test suite packaging all the tests for your application
1. A Docker image that runs code wrapping Kurtosis' controller code
1. A CLI that wraps Kurtosis' initializer code

More specifically, these are all the bits you'll need to implement in tutorial form:

### The Test Network
A test runs against a network of services that you're testing, and we'll need to tell Kurtosis what that network should look like. Networks are composed of services, so we'll first need to tell Kurtosis what a "service" looks like. 

In this example, we'll suppose that our network is composed of REST microservices running in Docker images, all communicating with each other, so we'll start by defining an interface that represents the functions a test can call on a node in our network. This is easily accomplished by implementing the [Service](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service.go) marker interface like so:

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

Now we have an interface that a test can use to interact with services in our network, but we still need to provide Kurtosis with an implementation. Every service in Kurtosis is a Docker container at time of writing, so our interface implementation will represent the actual service running in the Docker container. To fulfill the `MyService` contract we simply need to return the socket of the service, which means that our implementation just needs the IP address of the Docker container and the port it's listening on:

```go
type MyServiceImpl struct {
    IpAddr string
    Port nat.Port
}

func (s MyServiceImpl) GetHttpRestSocket() ServiceSocket {
    return ServiceSocket{IpAddr: s.IpAddr, Port: s.Port}
}
```

This way, our tests can use a nice interface for interacting with a service and we cleanly represent the reality that the service is backed by a Docker container with an IP address, listening on a port.

Next, we need to give Kurtosis the details on how to actually start a Docker container running one of our services. Kurtosis makes this easy - we only need to fill out the [ServiceInitializerCore](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service_initializer_core.go) interface. This interface is well-documented, so we'll use the documentation to write a service initializer core for our service like so:

```go
const (
    // Our Docker image is hardcoded to listen on port 80
    httpRestPort nat.Port = "80/tcp"
)

type MyServiceInitializerCore struct {
    // This is a parameter specific to our service
    // We could put more service-specific parameters here as needed
    MyServiceParam1 string
}

func (core MyServiceInitializerCore) GetUsedPorts() map[nat.Port]bool {
    return map[nat.Port]bool{httpRestPort: true}
}

func (core MyServiceInitializerCore) GetServiceFromIp(ipAddr string) Service {
    return MyServiceImpl{IpAddr: ipAddr, Port: httpRestPort}
}

func (core MyServiceInitializerCore) GetFilesToMount() map[string]bool {
    return map[string]bool{
        "our-config-file": true,
    }
}

func (core MyServiceInitializerCore) InitializeMountedFiles(mountedFiles map[string]*os.File, dependencies []Service) error {
    // Do some initialization of our config file as specified by the key we declared above
}

func (core MyServiceInitializerCore) GetTestVolumeMountpoint() string {
    // Return a filepath on our Docker image that's safe to mount the Kurtosis test volume on
}

func (core MyServiceInitializerCore) GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr string, dependencies []Service) ([]string, error) {
    // Use my knowledge of the Docker image to craft a command to run the Docker image with
    // This is where we'll use any service-specific params from above, including MyServiceParam1
}
```

Note the service "dependencies" that show up above. Kurtosis knows that some services will depend on others, and gives the developer the option to modify a service's files and start command based on other existing services in the network. We'll see how to declare these dependencies later.

Since we don't want to run a test against a network of services that are still starting up, the last piece we need for our service is a way to tell Kurtosis when the service is actually available (since the Docker container being started does not mean that the service is actually available). We'll use the [ServiceAvailabilityCheckerCore](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/services/service_availability_checker_core.go) like so:

```go
type MyServiceAvailabilityCheckerCore struct {}

func (core MyServiceAvailabilityCheckerCore) IsServiceUp(toCheck Service, dependencies []Service) bool {
    // Go doesn't have generics, so we need to cast to our service interface
    castedToCheck := toCheck.(MyService)

    httpRestSocket := castedToCheck.GetHttpRestSocket()
    
    // Use the socket to query some healthcheck endpoint to verify our service is up and return a bool appropriately
}

func (core MyServiceAvailabilityCheckerCore) GetTimeout() time.Duration {
    // This could also be parameterizable
    return 30 * time.Second
}
```

We're all set up to use this service in a network now... but of course, we still need to define what that network looks like. Kurtosis has a [ServiceNetwork](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/networks/service_network.go) object that represents the underlying state of the test network, but interacting with it is often too low-level for writing clean tests. To make writing tests as simple as possible, Kurtosis lets the developer define an arbitrary network struct that wraps the low-level network representation; this struct will then be passed to the tests. The developer can define this higher-level network wrapper object any way they please, but in our example we'll imagine that our tests all use a three-node network so we'll define the following:

```go
type ThreeNodeNetwork struct {
    networks.Network

    BootNode MyService
    ChildNode1 MyService
    ChildNode2 MyService
}
```

Much like `MyService`, this provides a test-friendly interface for interacting with the network. But, also like `MyService`, we need to define how to actually initialize this object. To do so, we bridge the gap between Kurtosis' `ServiceNetwork` and our custom `ThreeNodeNetwork` with an implementation of the [TestNetworkLoader](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_network_loader.go) interface.  

Before we write our implementation though, it's worth understanding how Kurtosis networks are configured. Each network has one or more **service configurations**, which are defined by a configuration ID, a docker image, a service initializer core, and an availability checker. These configurations serve as templates for the service instances that will eventually comprise the network; if a network is composed of only one type of service then the network only needs one configuration, but a network of many different types of services will need many configurations.

Using this information and the documentation on `TestNetworkLoader`, we can now write our `ThreeNodeNetworkLoader` implementation:

```go
const (
    configId = 0

    bootNodeServiceId = 0
    childNode1ServiceId = 1
    childNode2ServiceId = 2
)

type ThreeNodeNetworkLoader struct {
    DockerImage string
    MyServiceParam1 string
}

func (loader ThreeNodeNetworkLoader) ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error {
    initializerCore := MyServiceInitializerCore{MyServiceParam1: loader.MyServiceParam1}
    checkerCore := MyServiceAvailabilityCheckerCore{}

    // TODO when we simplify configuration-defining, change this name to match
    builder.AddStaticImageConfiguration(configId, loader.DockerImage, initializerCore, checkerCore)
    return nil
}

func (loader ThreeNodeNetworkLoader) InitializeNetwork(network *ServiceNetwork) (map[int]services.ServiceAvailabilityChecker, error) {
    result := map[int]services.ServiceAvailabilityChecker{}

    // Create the boot node using the configuration we defined earlier
    bootChecker, err := network.AddService(configId, bootNodeServiceId, map[int]bool{})
    // ... error-checking omitted ...
    result[bootNodeServiceId] = bootChecker

    // Define child nodes that depend on the boot node (Go doesn't have a set type, so a map[int]bool is used instead)
    // NOTE: Error-checking has been omitted
    childNode1Checker, err := network.AddService(configId, childNode1ServiceId, map[int]bool{bootNodeServiceId: true})
    result[childNode1ServiceId] = childNode1Checker
    childNode2Checker, err := network.AddService(configId, childNode2ServiceId, map[int]bool{bootNodeServiceId: true})
    result[childNode2ServiceId] = childNode2Checker

    return result, nil
}

func (loader ThreeNodeNetworkLoader) WrapNetwork(network *ServiceNetwork) (Network, error) {
    // By moving the interaction with the ServiceNetwork here, we remove the need for the test itself to know how to do this
    bootNodeService := network.GetService(bootNodeServiceId).Service.(MyService)
    childNode1Service := network.GetService(childNode1Service).Service.(MyService)
    childNode2Service := network.GetService(childNode2Service).Service.(MyService)

    return ThreeNodeNetwork{
        BootNode: bootNodeService,
        ChildNode1: childNode1Service,
        ChildNode2: childNode2Service,
    }
}
```

Here, we can see service dependencies in use: we have a boot node that depends on no other nodes (and so receives an empty dependency set), and two children node who depend on the boot node (and so declare a dependency set of the boot node service ID). 

The heavy lifting is finally done - we've declared a service with the appropriate initializer and availability checker cores, a network composed of that service, and a loader to wrap the low-level Kurtosis representation with a simpler, test-friendly version. Let's start writing some tests!


### The Test Suite
A test suite is simply a package of tests, and a test is just a definition of the required test network and a chunk of logic that validates against it. To write a test we'll need to implement the [Test](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test.go) interface like so:

```go
type ThreeNodeNetworkTest1 struct {
    DockerImage string
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
    return ThreeNodeNetworkLoader{DockerImage: test.DockerImage, MyServiceParam1: "some-test-specific-value"}
}

func (test ThreeNodeNetworkTest1) GetTimeout() time.Duration {
    return 30 * time.Second
}

```

Note that test failures are reported using the [TestContext](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_context.go) object, in a manner similar to Go's inbuilt `testing.T` object.

Now that we have a test, we can an implementation of the [TestSuite](https://github.com/kurtosis-tech/kurtosis/blob/develop/commons/testsuite/test_suite.go) interface to package it:

```go
type MyTestSuite struct {
    DockerImage string
}

func (suite MyTestSuite) GetTests() map[string]Test {
    return map[string]Test {
        "threeNodeNetworkTest1": ThreeNodeNetworkTest1{
            DockerImage: suite.DockerImage,
        },
    }
}
```

We're almost there - we just need to use our test suite!

### The Controller
To orchestrate all the steps required to run a single test, we need to provide Kurtosis a controller Docker image that will run code that instantiates our test suite, passes it to an instance of Kurtosis' [TestController](https://github.com/kurtosis-tech/kurtosis/blob/develop/controller/test_controller.go), and calls the `RunTest` function to run the test. This means that we need to write a main function that performs the steps above, and a Dockerfile that will generate an image to run our main function.

When we look at what we need to write in our main function, we discover that we already have our test suite but creating a new instance of a `TestController` requires many arguments that we won't know how to provide. Fortunately, the Kurtosis initializer will pass our controller Docker container these values via Docker environment variables. The complete list of the environment variables that our image will receive is defined in the `generateTestControllerEnvVariables` function inside [TestExecutor](https://github.com/kurtosis-tech/kurtosis/blob/develop/initializer/parallelism/test_executor.go), so we'll need to make sure that we receive them in our Dockerfile:

```
# ...image-specific Docker initialization things

# NOTE: Environment variables passed in as of 2020-07-19
CMD ./controller \
    --test=${TEST_NAME} \
    --subnet-mask=${SUBNET_MASK} \
    --docker-network=${NETWORK_NAME} \
    --gateway-ip=${GATEWAY_IP} \
    --log-level=${LOG_LEVEL} \
    # TODO refactor this when the service-config-definition cleanup happens
    --docker-image-name=${TEST_IMAGE_NAME} \
    --test-controller-ip=${TEST_CONTROLLER_IP} \
    --test-volume=${TEST_VOLUME} \
    --test-volume-mountpoint=${TEST_VOLUME_MOUNTPOINT} &> ${LOG_FILEPATH}
```

We'll then need to use these flags in our main function to create our `TestController` and return an exit code appropriate to the test result:

```go
func main() {
    testNameArg := flag.String("test", "", "The name of the test the controller will run")
    subnetMaskArg := flag.String("subnet-mask", "", "The name of the subnet the controller will run in")
    // ... etc....

    testSuite := MyTestSuite{DockerImage: *dockerImageNameArg}
    controller := controller.NewTestController(
        *testVolumeArg,
        *testVolumeMountpointArg,
        *dockerNetworkArg,
        *subnetMaskArg,
        *gatewayIpArg,
        *testControllerIpArg,
        testSuite,
        *testImageNameArg)

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

Once we build the Docker image, we'll have an image we can use with the initializer to run our test suite!

### The Initializer
We of course want to run our test suite via CI, which means we need a concrete entrypoint that our CI system can call to run the suite. We'll therefore need to build a main function to actually run our suite. Kurtosis makes this very simple - just write a main function that creates an instance of our test suite, plug it into an instance of [TestSuiteRunner](https://github.com/kurtosis-tech/kurtosis/blob/develop/initializer/test_suite_runner.go) along with the controller image to run, and have the CLI return an exit code corresponding to test results:

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
        *serviceImageNameArg,   // TODO remove this redundant declaration with the service-config-simplification refactor
        *controllerImagNameArg,
        map[string]string{},   // TODO provide an example with custom environment variables
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

And with this last bit, we're ready to run our test suite! Compiling and running our main function will 



TODO things ommitted: modifying the network dynamically during a test

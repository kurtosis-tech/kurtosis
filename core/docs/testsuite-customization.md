Testsuite Customization
=======================
### Prerequisites
* Bootstrapped testsuite from following the quickstart instructions [here](https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart)
* Testsuite opened in your IDE of choice for your testsuite's language

### Testsuite Customization
Now that you have a running testsuite, you'll want to start customizing the testsuite to your needs. Here we'll walk through the components inside a testsuite, as well as how to customize them.

**NOTE:** This tutorial will avoid language-specific idioms and use a Java-like pseudocode notation in this documentation, but will reference the example in [the Go implementation](https://github.com/kurtosis-tech/kurtosis-libs/tree/master/golang) to illustrate. All object names and methods will be more or less the same in your language of choice, and [all language repos](https://github.com/kurtosis-tech/kurtosis-libs/tree/master) come with an example testsuite.

### Service Interface & Implementation
You're writing a Kurtosis testsuite because you want to write tests for a network, and networks are composed of services. To a test, a service is just an API that represents an interaction with the actual Docker container. To give your tests this API, define an implementation of the `Service` interface that provides the functionality you want for your test [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/services_impl/datastore/datastore_service.go). Here you'll define all the functions you want to be able to call on the service, as well as the Kurtosis-required function `Service.isAvailable` for telling Kurtosis when your service is available and ready for use.

By now, you should have an implementation of the `Service` interface that represents an instance of your service.

### Container Config Factory
Our tests now have a nice interface for interacting with a service running in a Docker container, but we need to tell Kurtosis how to actually start the Docker container running the service. This is done by implementing the `ContainerConfigFactory` interface. [This interface is well-documented in the documentation](https://docs.kurtosistech.com/kurtosis-libs/lib-documentation#containerconfigfactorys-extends-service), so you can use the guidance there to write an initializer [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/develop/golang/testsuite/services_impl/datastore/datastore_container_config_factory.go).

### Tests & Test Setup
Now that we have a service, we can use it in a test. Each test is simply an implementation of the `Test` interface, and each has a `Test.setup` method which performs the work necessary to setup the testnet to a state where the test can run over it. You should use the `NetworkContext.addService` method to create instances of your service [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/basic_datastore_test/basic_datastore_test_.go#L38). 

The `addService` call return an instance of your service, as well as an `AvailabilityChecker` object. The `waitForStartup` method of the checker is a polling wrapper around `Service.isAvailable` that you wrote earlier, and can be used to block until the service instance is up or a timeout is hit.

Finally, the `Test.setup` method must return a `Network` object. This returned object will be the same one passed in as an argument to the `Test.run` method, which the test can use to interact with the network. For now, you can return the `NetworkContext` object.

### Test Logic
Every implementation of the `Test` interface must fill out the `Test.run` method. This function takes in the `Network` object that was returned by `Test.setup`, and uses the methods on the `TestContext` object to make assertions about the state of the network [like so](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/basic_datastore_test/basic_datastore_test_.go#L48). If no failures are called using the `TestContext`, the test is assumed to pass.

### Service Dependencies
It's not very useful to test just one service at a time; we're using Kurtosis because we want to test whole networks. This means that we need services which depend on other services. Fortunately, this is easily done by passing the dependency `Service` interface in the dependent's `ContainerConfigFactory` constructor like you would any other object, like [this API service which depends on the datastore service](https://github.com/kurtosis-tech/kurtosis-libs/blob/develop/golang/testsuite/services_impl/api/api_container_config_factory.go#L37).

Then, when instantiating the network in `Test.setup`, simply instantiate the dependency first and the dependent second, [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/basic_datastore_and_api_test/basic_datastore_and_api_test_.go#L39).

### Test Suite
Now that you have a test, you'll need to package it into a testsuite. A testsuite is simply an implementation of the `TestSuite` interface that yields a set of named tests, [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/example_testsuite.go). This is also where you'll thread through parameterization, like what Docker image the tests should run with.

### Test Suite Executor & Configurator
Finally, you'll need to tell Kurtosis how to initialize your testsuite. All test suites are run via the `TestSuiteExecutor` class, which is configured at construction time with an instance of the `TestSuiteConfigurator` interface. This configurator class is responsible for doing things like setting the log level and constructing the instance of your testsuite from custom params (more on these later), so you'll need to create your own implementation [like this](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/execution_impl/example_testsuite_configurator.go).

### Main Function
With your testsuite configurator complete, your only remaining step is to make sure it's getting used. When you bootstrapped your testsuite repo, you will have received an entrypoint main function that runs the testsuite executor [like this Go example](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/main.go). You will also have received a Dockerfile, for packaging that main CLI into a Docker image ([Go example](https://github.com/kurtosis-tech/kurtosis-libs/blob/develop/golang/testsuite/Dockerfile)). 

What `build-and-run.sh` actually does during its "build" phase is compile the main entrypoint CLI and package it into an executable Docker image. In order for your testsuite to get run, you just need to make sure this main entrypoint CLI is using your `TestsuiteConfigurator` by slotting in your configurator where indicated.

Congratulations - you now have your custom testsuite running using Kurtosis!

<!-- TODO MOVE THIS STUFF TO ADVANCED USAGE SECTION -->
### Custom Networks
So far your `Test.setup` method has returned the Kurtosis-provided `NetworkContext`, and your `Test.run` method has consumed it. This can be enough for basic tests, but you'll often want to centralize the network setup logic into a custom object that all your tests will use. Kurtosis allows this by letting your `Test.setup` method return any implementation of the `Network` marker interface; the `Test.run` will then receive that same `Network` object as an argument. To see this in action, the Go example testsuite has [this custom `Network` object](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/networks_impl/test_network.go), which makes the `Test.setup` of complex networks [a whole lot simpler](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/advanced_network_test/advanced_network_test_.go#L34) by encapsulating all the `ContainerConfigFactory` instantiation and waiting-for-availability.

### Custom Parameterization
You'll notice that the `TestsuiteConfigurator.parseParamsAndCreateSuite` method takes in a "params JSON" argument. This is arbitrary data that you can pass in to customize your testsuite's behaviour (e.g. which tests get run, or which Docker images get used). The data you pass in here is up to you, and is set via the `--custom-params` flag when calling `build-and-run.sh`. To see this in action, look at how [the example parses the args to a custom object that it uses to instantiate the testsuite](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/execution_impl/example_testsuite_configurator.go#L36).

### Visual Overview
To provide a visual recap of everything you've done, here's a diagram showing the control flow between components:

![](./images/testsuite-architecture.png)

---

[Back to index](https://docs.kurtosistech.com)

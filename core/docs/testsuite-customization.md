Testsuite Customization
=======================
### Prerequisites
* Bootstrapped testsuite from following the quickstart instructions [here](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/tree/master#testsuite-quickstart)
* Testsuite opened in your IDE of choice for your testsuite's language

### Testsuite Customization
Now that you have a running testsuite, you'll want to start customizing the testsuite to your needs. Here we'll walk through the components inside a testsuite, as well as how to customize them.

**NOTE:** This tutorial will avoid language-specific idioms and use a Java-like pseudocode notation in this documentation, but will reference the example in [the Go implementation](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/tree/master/golang) to illustrate. All object names and methods will be more or less the same in your language of choice, and [all language repos](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/tree/master) come with an example testsuite.

### Tests & Test Setup
First, we need to write a test. Each test is simply an implementation of the `Test` interface, and each has a `Test.setup` method which performs the work necessary to setup the testnet to a state where the test can run over it. You should use the `NetworkContext.addService` method to create instances of your service [like this](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/testsuite_impl/basic_datastore_test/basic_datastore_test_.go#L41). 

The `addService` call returns a `ServiceContext` object, which will contain the IP address of the service started so that you can wrap it in the appropriate client for your service (e.g. the Go Elasticsearch client). If your client doesn't have a way to check if the service is available, you can use the provided `NetworkContext.waitForAvailability` method as a convenience to wait until your service is available.

Finally, the `Test.setup` method must return a `Network` object. This returned object will be the same one passed in as an argument to the `Test.run` method, which the test can use to interact with the network. For now, you can return the `NetworkContext` object.

Go ahead and create your own `Test` implementation now, with a `Test.setup` method that sets up the network you'd like to test.

### Test Logic
Every implementation of the `Test` interface must fill out the `Test.run` method. This function takes in the `Network` object that was returned by `Test.setup`, and uses the methods on the `TestContext` object to make assertions about the state of the network [like so](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/testsuite_impl/basic_datastore_test/basic_datastore_test_.go#L48). If no failures are called using the `TestContext`, the test is assumed to pass.

You should now fill in your test's `run` method with logic to query and make assertions on your test network.

### Test Suite
Now that you have a test, you'll need to package it into a testsuite. A testsuite is simply an implementation of the `TestSuite` interface that yields a set of named tests, [like this](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/testsuite_impl/example_testsuite.go). This is also where you'll thread through parameterization, like what Docker image the tests should run with.

You can go ahead and create your own `TestSuite` implementation now, to contain your test.

### Test Suite Executor & Configurator
Finally, you'll need to tell Kurtosis how to initialize your testsuite. All test suites are run via the `TestSuiteExecutor` class, which is configured at construction time with an instance of the `TestSuiteConfigurator` interface. This configurator class is responsible for doing things like setting the log level and constructing the instance of your testsuite from custom params (more on these later), so you'll need to create your own implementation [like this](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/execution_impl/example_testsuite_configurator.go).

You should create your own `TestSuiteConfigurator` implementation now, to tell Kurtosis how to create your testsuite.

### Main Function
With your testsuite configurator complete, your only remaining step is to make sure it's getting used. When you bootstrapped your testsuite repo, you will have received an entrypoint main function that runs the testsuite executor [like this Go example](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/main.go). You will also have received a Dockerfile, for packaging that main CLI into a Docker image ([Go example](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/develop/golang/testsuite/Dockerfile)). 

What `build-and-run.sh` actually does during its "build" phase is compile the main entrypoint CLI and package it into an executable Docker image. In order for your testsuite to get run, you just need to make sure this main entrypoint CLI is using your `TestsuiteConfigurator` by slotting in your configurator where indicated.

Congratulations - you now have your custom testsuite running using Kurtosis!

<!-- TODO MOVE THIS STUFF TO ADVANCED USAGE SECTION -->
### Custom Networks
So far your `Test.setup` method has returned the Kurtosis-provided `NetworkContext`, and your `Test.run` method has consumed it. This can be enough for basic tests, but you'll often want to centralize the network setup logic into a custom object that all your tests will use. Kurtosis allows this by letting your `Test.setup` method return any implementation of the `Network` marker interface; the `Test.run` will then receive that same `Network` object as an argument. To see this in action, the Go example testsuite has [this custom `Network` object](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/networks_impl/test_network.go), which makes the `Test.setup` of complex networks [a whole lot simpler](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/testsuite_impl/advanced_network_test/advanced_network_test_.go#L36) by encapsulating all the container-starting and availability-checking.

If you'd like, you can extract your test's `Test.setup` logic into a custom `Network` implementation to make your custom test code cleaner.

### Custom Parameterization
You'll notice that the `TestsuiteConfigurator.parseParamsAndCreateSuite` method takes in a "params JSON" argument. This is arbitrary data that you can pass in to customize your testsuite's behaviour (e.g. which tests get run, or which Docker images get used). The data you pass in here is up to you, and is set via the `--custom-params` flag when calling `build-and-run.sh`. To see this in action, look at how [the example parses the args to a custom object that it uses to instantiate the testsuite](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/golang/testsuite/execution_impl/example_testsuite_configurator.go#L36).

If your testsuite needs custom parameters (e.g. the name of the Docker images to your test network), you can parameterize your `TestSuite` implementation and consume them in your custom executor.

### Visual Overview
To provide a visual recap of everything you've done, here's a diagram showing the control flow between components:

![](./images/testsuite-architecture.png)

---

[Back to index](https://docs.kurtosistech.com)

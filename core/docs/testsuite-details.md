Testsuite
=========
Overview
--------
Every testsuite is simply a Docker image that runs a CLI that executes a test. The test execution logic for the CLI is encapsulated in [a Kurtosis client library in your language of choice](https://github.com/kurtosis-tech/kurtosis-docs/blob/master/supported-languages.md), so your main CLI function will only be a thin layer of arg-parsing, an instantiation of your suite object that contains test and testnet info, and a call to the Kurtosis client's `Client.Run` function. The components required to instantiate a testsuite are as follows:

![](./images/testsuite-architecture.png)

Each client library contains an example implementation of the components as well as the ability to bootstrap a new testsuite from the example implementation. The rest of the guide will assume you've bootstrapped a new testsuite, so if you haven't done so already we recommend following the [the quickstart instructions](./quickstart.md) to bootstrap now.

**NOTES:**
* Kurtosis provides clients for writing testsuites in multiple languages. While the languages differ, the objects and function calls are named the same. For consistency, this guide will avoid language-specific idioms and will use pseudocode with a generic `Object.function` notation.
* Each Kurtosis client provides comments on all objects and functions, so this guide will focus on the higher-level interaction between the components rather than specific documentation of each function and argument. For detailed docs, see the in-code comments in your client of choice.

Components
----------
### CLI Main Function
Your Kurtosis testsuite CLI in your repo will be a `main` function that does the following in order:

1. **Arg-Parsing:** Your CLI will receive several Kurtosis-specific arguments, which (with the exception of log level) will in turn be passed as-is to the `Client.run` function. You can also receive custom arguments specific to your testsuite (more on this later).
2. **Set Log Level:** Using the log level arg, set the logging level for the test's execution.
3. **TestSuite Instantiation:** The `Client.run` function needs details about the test logic to run as well as instructions on setting up the network required. These are contained in a `TestSuite` object, whose behaviour can be modified with custom arguments specific to your testsuite.
4. **Test Execution:** Calling `Client.run` with the Kurtosis arguments and `TestSuite` object.

### Dockerfile
To package the CLI into a Docker image, your repo will have a Dockerfile under the example implementation folder that defines how to build the image (and if Dockerfiles are alien to you, we recommend [the official Docker docs](https://docs.docker.com/get-started/) as a great place to start). Kurtosis testsuite Dockerfiles are very simple, and simply compile and run the CLI. 

The only bit of complexity is that the Dockerfile will receive Kurtosis-specific parameters as magic environment variables, which will then be passed to your testsuite CLI in the form of flag args. These environment variables are as follows:

* `DEBUGGER_PORT`
* `KURTOSIS_API_IP`
* `LOG_LEVEL`
* `METADATA_FILEPATH`
* `SERVICES_RELATIVE_DIRPATH`
* `TEST`

With the exception of the log level and debugger port, all of these will be passed as-is to the `Client.run` call. If you modify the Dockerfile, you will need to make sure that you continue to receive these variables as flags in your CLI main function.

### TestSuite
Every Kurtosis client's `Client.Run` function requires a `TestSuite` object that contains details about:

* Your services - what Docker images to use, what params to give when running the container, how to verify the service is available, etc.
* Your testnets - their topology and what types and quantities of services they're composed of
* Your tests - their names and logic, what testnets they want, their timeouts, etc.

All this information is packaged inside the `Test` object, so a `TestSuite` is really a wrapper class for a set of named `Test`s.

### Test
A `Test` object packages the logic for executing an individual test with a definition of the network the test will execute against. It has three main components:

* A `setup` method for creating the testnet that the test will run against
* A `run` function that takes in a `TestContext` for making assertions and a `Network` object that serves as a handle to interacting with the testnet
* Timeouts defining when the test should be marked as failed due to running for too long

### Network
A `Network` object is an optional, entirely user-defined representation of a running testnet wrapped around a `NetworkContext`. Its purpose is to provide a layer of abstraction so users can make test-writing as simple as possible. For example, if all your tests want a five-node network, you could write a `FiveNodeNetwork` object that implements the `Network` interface and give it functions like `getNodeOne`, `getNodeTwo`, etc. 

### DockerContainerInitializer
A `DockerContainerInitializer` provides specific information about how to launch the actual Docker container underlying a given `Service`, for use when defining service configurations in the `NetworkLoader`. This object is very well-documented in code, so we recommend users read the details in their client library of choice.

### Service
The `Service` interface represents a service running in a Docker container. For example, an `ElasticsearchService` implementation might have a `getClient` function that returns an ES client so a test can easily interact with an Elasticsearch service.

Running A Testsuite
-------------------
As detailed in [the architecture docs](./architecture.md), testsuite containers are launched, one per test, by the Kurtosis initializer container. The initializer container itself is a sort of "CLI", and the entrypoint into running the Kurtosis platform; however, launching it is nontrivial as it requires several special flags to its `docker run` command. Further, because the initializer is a CLI, it receives its own flags (separate from `docker run`'s flags) to customize its behavior!

This can become very confusing very quickly, so every bootstrapped repo comes with a `build_and_run.sh` script in the `scripts` directory to make building and running your testsuite simple. Run `build_and_run.sh help` inside your bootstrapped repo to display detailed information about how to use this script.

As the script's help text mentions, the execution of the testsuite can be modified by adding additional Docker `--env` flags to set certain Kurtosis environment variables. `build_and_run.sh` already sets all the required parameters by default, but if you want to customize execution further you can see all available parameters to the initializer with the `SHOW_HELP` Docker environment like so: `build_and_run.sh all --env SHOW_HELP=true`.

Customizing Testsuite Execution
-------------------------------
You'll very likely want to customize the behaviour of your testsuite based on information passed in when you execute Kurtosis (e.g. have a `--fast-tests-only` flag to your CLI's main function that runs a subset of the tests in your suite). To do so, you'll need to:

1. Add the appropriate flags to your CLI's main function
1. Edit the `Dockerfile` that wraps your testsuite CLI to set the flag using a Docker environment variable (e.g. `--fast-tests-only=${FAST_TESTS_ONLY}`)
1. When running Kurtosis, use the `--env` Docker flag to set the value of the initializer's `CUSTOM_ENV_VARS_JSON` parameter-variable to a JSON map containing your desired values (e.g. `build_and_run.sh all --env CUSTOM_ENV_VARS_JSON="{\"FAST_TESTS_ONLY\":true}"`)

**WARNING:** Docker doesn't like unescaped spaces when using the `--env` flag. You'll therefore want to backslash-escape BOTH spaces and double-quotes, like so: `--env CUSTOM_ENV_VARS_JSON="{\"VAR1\":\ \"var1-value\",\ \"VAR2\":\ \"var2-value\"}"`.

Next Steps
----------
Now that you understand more about the internals of a testsuite, you can:

* Head over to [the quickstart instructions](./quickstart.md) to bootstrap your own testsuite (if you haven't already)
* Visit [the architecture docs](./architecture.md) to learn more about the Kurtosis platform at a high level.
* Check out [the instructions for running in CI](./running-in-ci.md) to see what's necessary to get Kurtosis running in your CI environment
* Pop into [the Kurtosis Discord](https://discord.gg/6Jjp9c89z9) to join the community!

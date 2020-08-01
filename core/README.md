Kurtosis
========
Kurtosis is a framework on top of Docker for writing test suites for any networked system - be it blockchain, distributed datastore, or otherwise. It handles all the gruntwork of setup, test execution, and teardown so you don't have to.

* [Architecture](#architecture)
* [Tutorial](#tutorial)
* [Examples](#examples)
* [Notes](#notes)
    * [Abnormal Exit](#abnormal-exit)
    * [Container, Volume, &amp; Image Tidying](#container-volume--image-tidying)

Getting Started
---------------
Kurtosis is a testing framework atop Docker, meaning you'll need to build a Docker image of the service(s) you want to test, and a Golang implementation of several Kurtosis interfaces to construct your test suite. If you've never used:
* Go, we recommend the [official Golang installation & quickstart guide](https://golang.org/doc/install) along with [JetBrains' excellent GoLand IDE](https://www.jetbrains.com/go/)
* Docker, we recommend [the official Docker "Get Started" guide](https://docs.docker.com/get-started/), coupled with [the docs explaining how to view container logs](https://docs.docker.com/config/containers/logging/) (which you'll be doing a lot)

### Architecture
The Kurtosis architecture has four components:

1. The **test networks**, which are the networks (one network per test) of service containers that are spun up for tests to run against
1. The **test suite**, which contains the package of tests that can be run
1. The **controller**, which is the Docker image that will be used to launch a container (one per test) responsible for orchestrating the execution of a single test
1. The **initializer**, which is the entrypoint to the testing application for CI

The control flow goes:

1. The initializer launches and looks at what tests need to be run
1. In parallel, for each test:
    1. The initializer launches a controller Docker container to orchestrate execution of the particular test
    1. The controller spins up a network of whichever Docker services the test requires
    1. The controller waits for the network to become available
    1. The controller runs the logic of the test it's assigned to run
    1. After the test finishes, the controller tears down the network of services that the test was using
    1. The controller returns the result to the initializer and exits
1. The initializer waits for all tests to complete and returns the results

### Building An Implementation
See [the "Getting Started" tutorial](./tutorials/getting-started.md) for a step-by-step tutorial on how to build a Kurtosis implementation from scratch.

### Debugging Failed Tests
See [the "Debugging Failed Tests" tutorial](./tutorials/debugging-failed-tests.md) for information on how to approach some common failure scenarios.

Examples
--------
See [the Ava end-to-end tests](https://github.com/kurtosis-tech/ava-e2e-tests) for the reference Kurtosis implementation.

Developer Notes
---------------
### Docker-in-Docker & MacOS Users
**High-level:** If you're using MacOS, make sure that your Docker engine's `Resources > File Sharing` preferences are set to allow `/var/folders`
**Details:** The Kurtosis controller is a Docker image that needs to access the Docker engine it's running in to create other Docker images. This is done via creating "sibling" containers, as detailed in the "Solution" section at the bottom of [this blog post](https://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/). However, this requires your Docker engine's communication socket to be bind-mounted inside the controller container. Kurtosis will do this for you, but you'll need to give Docker permission for the Docker socket (which lives at `/var/run/docker.sock`) to be bind-mounted inside the controller container.

### Parallelism
Kurtosis offers the ability to run tests in parallel to reduce total test suite runtime. You should never set parallelism higher than the number of cores on your machine or else you'll actually slow down your tests as your machine is doing unnecessary context-switching; depending on your test timeouts, this could cause spurious test failures.

### Abnormal Exit
While running, Kurtosis will create the following, per test:
* A new Docker network for the test
* A new Docker volume to pass files relevant to the test in
* Several containers related to the test

**If Kurtosis is killed abnormally (e.g. SIGKILL or SIGQUIT), the user will need to remove the Docker network and stop the running containers!** The specifics will depend on what Docker containers you start, but the network and container cleanup can be done using something similar to the following:

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

### Container, Volume, & Image Tidying
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

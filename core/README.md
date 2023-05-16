Kurtosis
========
Kurtosis is a framework on top of Docker for writing test suites for any networked system - be it blockchain, distributed datastore, or otherwise. It handles all the gruntwork of setup, test execution, and teardown so you don't have to.

Official docs found [here](https://docs.kurtosis.com) (created using Docusaurus from this repo [here](https://github.com/kurtosis-tech/kurtosis/tree/main/docs).

Development Prerequisites
-------------------------
* Docker engine installed on your machine
* `protoc` installed (can be installed on Mac with `brew install protobuf`)
* The Golang extension to `protoc` installed (can be installed on Mac with `brew install protoc-gen-go`)
* The Golang gRPC extension to `protoc` installed (can be installed on Mac with `brew install protoc-gen-go-grpc`)

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

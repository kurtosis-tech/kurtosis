Kurtosis With Docker
====================
Kurtosis uses Docker for its ability to run tests in a lightweight, reproducible fashion. Here we'll look at what Kurtosis does on the Docker level so you get a better understanding of what's going on under the hood. 

Prerequisites
-------------
* [Getting Started](./getting-started.md)

Tutorial
--------
Kurtosis' workflow looks as follows, with interactions with the Docker engine **highlighted**:

1. The initializer launches and looks at what tests need to run
1. For each test to run:
    1. **The initializer creates a new [Docker network](https://docs.docker.com/network/) to house the containers for the test**
    1. **The initializer creates a new [Docker volume](https://docs.docker.com/storage/volumes/) as a shared location to store test files (including logfiles)**
    1. **The initializer launches a Docker container running the test controller image, with the test volume mounted and available to the controller**
    1. The controller container starts up, writing logs to the test volume
    1. The controller looks at what services are required to run the test network needed to run its assigned test and, for each service:
        1. Creates a directory in the test volume for the service being started
        1. Initializes any files that the service asks for in the directory
        1. **Launches a container running the image requested for the service, with the test volume mounted**
        1. Waits for the service to be available as defined using the logic in the service's availability checker
    1. The controller runs the test logic, passing in information about the network of services that was started
    1. **On test completion, the controller instructs each container in the test network to stop and waits for their exit**
    1. The controller container exits with a status code corresponding to whether the test passed or failed
    1. **The initializer, who has been waiting for the test controller container to finish, receives the controller exit code**
    1. The initializer prints the controller's logs from the test volume
    1. **The initializer tears down the Docker network that the test network was running in**
1. The initializer prints a summary of all test statuses
1. If any test failed, the initializer exits with a non-zero exit code

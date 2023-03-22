
<img src="./logo.png" width="500">

----

[Kurtosis](https://www.kurtosis.com) is a build system for multi-container test environments.

Use cases for Kurtosis include:

- Running a third-party distributed app, without knowing how to set it up
- Local prototyping & development on distributed apps
- Writing integration and end-to-end distributed app tests (e.g. happy path & sad path tests, load tests, performance tests, etc.)
- Running integration/E2E distributed app tests
- Debugging distributed apps during development

## Why Kurtosis?

Kurtosis makes it much easier for a user to spin up multi-container test environments than alternatives like Helm Charts or Docker Compose; by working at a levle higher
than a container orchestrator. It gives super powers to anyone who wants to spin up multi-container test environments by allowing - 

- Instantiation of a multi container environment
    - Kurtosis allows you to mount files & dynamically generated data in the right places within different containers across different test runs on different platforms
    - Kurtosis allows you to refer to IP addresses, ports and hostnames of a service from a different service in the same multi container environment
    - Kurtosis allows you to write reproducible "wait" logic to make sure that the right services are up and healthy in the right way across test environment setups
    - Kurtosis is parametrized, allowing you to to tweak the number of nodes a service spins up with through user passed arguments
    - Kurtosis allows you to include other environments in your environment reducing the need to repeat yourself
    - Kurtosis allows you to port test environments, including the data setup, across platforms; allowing you to run the same test environment on your laptop, CI Job or an ad-hoc scale test environment
- Interaction with a running multi container environment
  - Kurtosis allows you to run on-box CLI commands within containers
  - Kurtosis allows you to send REST requests to containers
- Debugging of multi container environment
  - Kurtosis provides easy access to logs of a given service

---

## To start using Kurtosis

### Prerequisites

Docker must be installed and running on your machine:

```bash
docker version
```

If it's not, follow the instructions from the [Docker docs](https://docs.docker.com/get-docker/).

### Installing Kurtosis

On MacOS:

```bash
brew install kurtosis-tech/tap/kurtosis-cli
```

On Linux (apt):
```bash
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```

On Linux (yum):
```bash
echo '[kurtosis]
name=Kurtosis
baseurl=https://yum.fury.io/kurtosis-tech/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/kurtosis.repo
sudo yum install kurtosis-cli
```

### Running

First of all we can create [a simple Starlark script][starlark-explanation] to spin up multiple replicas of `httpd`:

```python
cat > script.star << EOF
def run(plan, args):
    configs = {}
    for i in range(args.replica_count):
       plan.add_service(
          "httpd-replica-"+str(i),
          config = ServiceConfig(
              image = "httpd",
              ports = {
                  "http": PortSpec(number = 8080),
              },
          ),
      )
EOF
```

Running the script gives us an enclave with three services:

```bash
kurtosis run script.star '{"replica_count": 3}'
```
```console
INFO[2023-03-22T13:27:03+01:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-22T13:27:07+01:00] Enclave 'arid-delta' created successfully

> add_service service_name="httpd-replica-0" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080)})
Service 'httpd-replica-0' added with service UUID '80030157c58f4eb2b9c98f41dd938ed0'

> add_service service_name="httpd-replica-1" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080)})
Service 'httpd-replica-1' added with service UUID '4abff039bfa74b019f0a0d2e155c760a'

> add_service service_name="httpd-replica-2" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080)})
Service 'httpd-replica-2' added with service UUID 'b6f90bc6dad748a4807ad4fbc3e5cc9a'

Starlark code successfully run. No output was returned.
INFO[2023-03-22T13:27:37+01:00] ===================================================
INFO[2023-03-22T13:27:37+01:00] ||          Created enclave: arid-delta          ||
INFO[2023-03-22T13:27:37+01:00] ===================================================
Name:            arid-delta
UUID:            0b3879d30c80
Status:          RUNNING
Creation Time:   Wed, 22 Mar 2023 13:27:03 CET

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID           Name              Ports                               Status
80030157c58f   httpd-replica-0   http: 8080/tcp -> 127.0.0.1:54164   RUNNING
4abff039bfa7   httpd-replica-1   http: 8080/tcp -> 127.0.0.1:54170   RUNNING
b6f90bc6dad7   httpd-replica-2   http: 8080/tcp -> 127.0.0.1:54174   RUNNING
```

### More examples

See our documentation on https://docs.kurtosis.com.

## To start contributing to Kurtosis

See our [CONTRIBUTING](./CONTRIBUTING.md) file.

--- 

## Repository Structure

This repository is structured as a monorepo, containing the following projects:
- `container-engine-lib`: Library used to abstract away container engine being used by the [enclave][enclave].
- `core`: Container launched inside an [enclave][enclave] to coordinate its state
- `engine`: Container launched to coordinate [enclaves][enclave]
- `api`: Defines the API of the Kurtosis platform (`engine` and `core`)
- `cli`: Produces CLI binary, allowing iteraction with the Kurtosis system
- `docs`: Documentation that is published to [docs.kurtosis.com](docs)
- `internal_testsuites`: End to end tests

## Dependencies

To build Kurtosis, you must the following dependencies installed:

- Bash (5 or above) + Git

On MacOS:
```bash
# Install modern version of bash, the one that ships on MacOS is too old
brew install bash
# Allow it as shell
echo "${BREW_PREFIX}/bin/bash" | sudo tee -a /etc/shells
# Optional: make bash your default shell
chsh -s "${BREW_PREFIX}/bin/bash"
# Install modern version of git, the one that ships on MacOS is too old
brew install git
```
- Docker
  
On MacOS:
```bash
brew install docker
```

- Go (1.18 or above)

On MacOS:
```bash
brew install go@1.18
```

- Goreleaser

On MacOS:
```bash
brew install goreleaser/tap/goreleaser
```

- Node (16.14 or above) and Yarn

On MacOS, using `NVM`:
```bash
brew install nvm
mkdir ~/.nvm
nvm install 16.14.0
npm install -g yarn
```
- Go and Typescript protobuf compiler binaries

On MacOS:
```bash
brew install protoc-gen-go
brew install protoc-gen-go-grpc
brew install protoc-gen-grpc-web
npm install -g ts-protoc-gen
npm install -g grpc-tools
```
- Musl

On MacOS:
```bash
brew install filosottile/musl-cross/musl-cross
```

## Unit Test Instructions

For all Go modules, run `go test ./...` on the module folder. For example:

```console
$ cd cli/cli/
$ go test ./...
```

## Build Instructions

To build the entire project, run:

```console
$ ./script/build.sh
```

To only build a specific project, run the script on `./PROJECT/PATH/script/build.sh`, for example:

```console
$ ./container-engine-lib/build.sh
$ ./core/scripts/build.sh
$ ./api/scripts/build.sh
$ ./engine/scripts/build.sh
$ ./cli/scripts/build.sh
```

If there are any changes to Protobuf files in `api`, the bindings must be regenerated:

```console
$ ./api/scripts/regenerate-protobuf-bindings.sh
```

Build scripts also run unit tests as part of the build proccess

## E2E Test Instructions

Each project's build script also runs the unit tests inside the project. Running `./script/build.sh` will guarantee that all unit tests in the monorepo pass.

To run the end to end tests:

1. Make sure Docker is running
```console
$ docker --version
Docker version X.Y.Z
```

2. Make sure Kurtosis Engine is running
```console
$ kurtosis engine status
A Kurtosis engine is running with the following info:
Version:   0.X.Y
```

1. Run `test.sh` script
```console
$ ./internal_testsuites/scripts/test.sh
```

If you are developing the Typescript test, make sure that you have first built `api/typescript`. Any
changes made to the Typescript package within `api/typescript` aren't hot loaded as of 2022-09-29.

## Run Instructions

To interact with the built CLI, run `./cli/cli/scripts/launch_cli.sh` as if it was the `kurtosis` command:

```console
$ ./cli/cli/scripts/launch_cli.sh enclave add
```

If you want tab completion on the recently built CLI, you can alias it to `kurtosis`:

```console
$ alias kurtosis="$(pwd)/cli/cli/scripts/launch_cli.sh"
$ kurtosis enclave add
```

## Questions, help, or feedback

If you have feedback for us or a question around how Kurtosis works and the [docs](https://docs.kurtosis.com) aren't enough, we're more than happy to help and chat about your use case via the following ways:
- Get help in our [Discord server](https://discord.gg/Es7QHbY4)
- Email us at [feedback@kurtosistech.com](mailto:feedback@kurtosistech.com)
- Schedule a 1:1 session with us [here](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding)

<!-------- ONLY LINKS BELOW THIS POINT -------->
[enclave]: https://docs.kurtosis.com/explanations/architecture#enclaves
[docs]: https://docs.kurtosis.com
[starlark-explanation]: https://docs.kurtosis.com/explanations/starlark


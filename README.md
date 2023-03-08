
<img src="./logo.png" width="500">

----

[Kurtosis](https://www.kurtosis.com) is a platform for orchestrating distributed system environments, allowing easy creation and manipulation of stage-appropriate deployments across the early stages of the development cycle (prototyping, testing).

Use cases for Kurtosis include:

- Enable individual developers to prototype on personal development environments without bothering with environment setup and configuration
- Enable development teams to run automated end-to-end tests for distributed systems, including fault tolerance tests in servers and networks, load testing, performance testing, etc.
- Enable developers to easily debug failing distributed systems during development

## Why Kurtosis?

Container management and container orchestration systems like Docker and Kubernetes are each great at serving developers in different parts of the development cycle (development for Docker, production for Kubernetes). These, and other distributed system deployment tools, are low-level, stage-specific tools that require teams of DevOps engineers to manage.

Kurtosis is designed to optimize environment management and control across the development cycle - operating at one level of abstraction higher than existing tools, giving developers the environments and the ability to manipulate them as needed at each stage.

---

## To start using Kurtosis

### Pre Requisite

Make sure Docker must be installed and running on your machine:
```bash
docker version
```

If you don't, follow the instructions from [Docker docs](https://docs.docker.com/get-docker/).

### Installing

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

First of all we can create a simple Starlark script to spin up multiple replicas of `httpd`:
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
WARN[2023-02-10T13:32:32-03:00] You are running an old version of the Kurtosis CLI; we suggest you to update it to the latest version, '0.66.3'
WARN[2023-02-10T13:32:32-03:00] You can manually upgrade the CLI tool following these instructions: https://docs.kurtosis.com/install#upgrading
INFO[2023-02-10T13:32:32-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-02-10T13:32:37-03:00] Enclave 'misty-bird' created successfully

> add_service service_name="httpd-replica-0" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080, transport_protocol="TCP", application_protocol="")})
Service 'httpd-replica-0' added with service UUID 'a05405bfe100475fa52883c71bd899e6'

> add_service service_name="httpd-replica-1" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080, transport_protocol="TCP", application_protocol="")})
Service 'httpd-replica-1' added with service UUID 'cc871310223c4bc7a539a6c93a8e33ea'

> add_service service_name="httpd-replica-2" config=ServiceConfig(image="httpd", ports={"http": PortSpec(number=8080, transport_protocol="TCP", application_protocol="")})
Service 'httpd-replica-2' added with service UUID 'dcfd1fb7a94e4e8e8a55c715f7f09b04'
Starlark code successfully run. No output was returned.
INFO[2023-02-10T13:32:53-03:00] ===================================================
INFO[2023-02-10T13:32:53-03:00] ||          Created enclave: misty-bird          ||
INFO[2023-02-10T13:32:53-03:00] ===================================================
```

That can be inspected by running

```console
kurtosis enclave inspect misty-bird
```
```console
WARN[2023-02-10T13:33:19-03:00] You are running an old version of the Kurtosis CLI; we suggest you to update it to the latest version, '0.66.3'
WARN[2023-02-10T13:33:19-03:00] You can manually upgrade the CLI tool following these instructions: https://docs.kurtosis.com/install#upgrading
UUID:                                 eeb28363fc53
Enclave Name:                         misty-bird
Enclave Status:                       RUNNING
Creation Time:                        Fri, 10 Feb 2023 13:32:32 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:63747
API Container Host GRPC Proxy Port:   127.0.0.1:63748

========================================== User Services ==========================================
UUID           Name              Ports                               Status
a05405bfe100   httpd-replica-0   http: 8080/tcp -> 127.0.0.1:63768   RUNNING
cc871310223c   httpd-replica-1   http: 8080/tcp -> 127.0.0.1:63772   RUNNING
dcfd1fb7a94e   httpd-replica-2   http: 8080/tcp -> 127.0.0.1:63781   RUNNING
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

- [Kudet](https://github.com/kurtosis-tech/kudet) (Kurtosis CLI helper)

On MacOS:
```bash
brew install kurtosis-tech/tap/kudet
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


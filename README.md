
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

- Bash
- Docker
- Go (1.18 or above)
- Node (16.14 or above)
- Yarn
- Protobuf binaries to generate bindings:
  * Go Protobuf compiler binaries (installable via OS package manager/Brew): `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-web`
  * Typescript Protobuf compiler binaries (installable via `npm install -g`: `ts-protoc-gen`, `grpc-tools`

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

<!-------- ONLY LINKS BELOW THIS POINT -------->
[enclave]: https://docs.kurtosis.com/explanations/architecture#enclaves
[docs]: https://docs.kurtosis.com


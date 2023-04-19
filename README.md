
<img src="./logo.png" width="1200">

----
What is Kurtosis?
=================
[Kurtosis](https://www.kurtosis.com) is a composable build system for multi-container test environments. Kurtosis makes it easier for developers to set up test environments that require dynamic setup logic (e.g. passing IPs or runtime-generated data between services) or programmatic data seeding. 

To read more about "why Kurtosis?", go [here](https://docs.kurtosis.com/#why-use-kurtosis).

To read about the architecture, go [here](https://docs.kurtosis.com/explanations/architecture).


Running Kurtosis
================

### Install

Follow the instructions [here](https://docs.kurtosis.com/install).

### Run
Kurtosis create ephemeral multi-container environments called [enclaves][enclave] using [Starlark](https://docs.kurtosis.com/concepts-reference/starlark). These can be bundled together into [packages](https://docs.kurtosis.com/concepts-reference/packages). Let's run one now:

```bash
kurtosis run github.com/kurtosis-tech/awesome-kurtosis/redis-voting-app
```

```console
INFO[2023-03-28T15:27:31-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-28T15:27:34-03:00] Enclave 'nameless-fjord' created successfully

> print msg="Spinning up the Redis Package"
Spinning up the Redis Package

> add_service service_name="redis" config=ServiceConfig(image="redis:alpine", ports={"client": PortSpec(number=6379, transport_protocol="TCP")})
Service 'redis' added with service UUID '3ca8b4c1c8344b2c96be1b988ba12a02'

> add_service service_name="voting-app" config=ServiceConfig(image="mcr.microsoft.com/azuredocs/azure-vote-front:v1", ports={"http": PortSpec(number=80, transport_protocol="TCP")}, env_vars={"REDIS": "{{kurtosis:5045f2098e4846b88efefbf2689b5538:hostname.runtime_value}}"})
Service 'voting-app' added with service UUID '8a8c18860df1440ca7bdc96fd511fb2a'

Starlark code successfully run. No output was returned.
INFO[2023-03-28T15:28:08-03:00] =======================================================
INFO[2023-03-28T15:28:08-03:00] ||          Created enclave: nameless-fjord          ||
INFO[2023-03-28T15:28:08-03:00] =======================================================
Name:            nameless-fjord
UUID:            6babc3090ad0
Status:          RUNNING
Creation Time:   Tue, 28 Mar 2023 15:27:31 -03

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID           Name         Ports                                 Status
3ca8b4c1c834   redis        client: 6379/tcp -> 127.0.0.1:58508   RUNNING
8a8c18860df1   voting-app   http: 80/tcp -> 127.0.0.1:58511       RUNNING
```

If this piqued your interest, you might like our [quickstart][quickstart-reference].

### More Examples

Further examples can be found in our [`awesome-kurtosis` repo][awesome-kurtosis].

Contributing to Kurtosis
========================

<details>
<summary>Expand to see contribution info</summary>

See our [CONTRIBUTING](./CONTRIBUTING.md) file.

Repository Structure
--------------------

This repository is structured as a monorepo, containing the following projects:
- `container-engine-lib`: Library used to abstract away container engine being used by the [enclave][enclave].
- `core`: Container launched inside an [enclave][enclave] to coordinate its state
- `engine`: Container launched to coordinate [enclaves][enclave]
- `api`: Defines the API of the Kurtosis platform (`engine` and `core`)
- `cli`: Produces CLI binary, allowing interaction with the Kurtosis system
- `docs`: Documentation that is published to [docs.kurtosis.com](docs)
- `internal_testsuites`: End to end tests

Dev Dependencies
----------------

To build Kurtosis itself, you must have the following installed:

#### Bash (5 or above) + Git

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
#### Docker

On MacOS:
```bash
brew install docker
```

#### Go (1.18 or above)

On MacOS:
```bash
brew install go@1.18
```

#### Goreleaser

On MacOS:
```bash
brew install goreleaser/tap/goreleaser
```

#### Node (16.14 or above) and Yarn

On MacOS, using `NVM`:
```bash
brew install nvm
mkdir ~/.nvm
nvm install 16.14.0
npm install -g yarn
```
#### Go and Typescript protobuf compiler binaries

On MacOS:
```bash
brew install protoc-gen-go
brew install protoc-gen-go-grpc
brew install protoc-gen-grpc-web
npm install -g ts-protoc-gen
npm install -g grpc-tools
```
#### Musl

On MacOS:
```bash
brew install filosottile/musl-cross/musl-cross
```

Build Instructions
------------------

To build the entire project, run:

```bash
./script/build.sh
```

To only build a specific project, run the script on `./PROJECT/PATH/script/build.sh`, for example:

```bash
./container-engine-lib/build.sh
./core/scripts/build.sh
./api/scripts/build.sh
./engine/scripts/build.sh
./cli/scripts/build.sh
```

If there are any changes to the Protobuf files in the `api` subdirectory, the Protobuf bindings must be regenerated:

```bash
$ ./api/scripts/regenerate-protobuf-bindings.sh
```

Build scripts also run unit tests as part of the build process.

Unit Test Instructions
----------------------

For all Go modules, run `go test ./...` on the module folder. For example:

```bash
cd cli/cli/
go test ./...
```

E2E Test Instructions
---------------------

Each project's build script also runs the unit tests inside the project. Running `./script/build.sh` will guarantee that all unit tests in the monorepo pass.

To run the end-to-end tests:

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

Dev Run Instructions
--------------------

Once the project has built, run `./cli/cli/scripts/launch_cli.sh` as if it was the `kurtosis` command:

```bash
./cli/cli/scripts/launch_cli.sh enclave add
```

If you want tab completion on the recently built CLI, you can alias it to `kurtosis`:

```bash
alias kurtosis="$(pwd)/cli/cli/scripts/launch_cli.sh"
kurtosis enclave add
```

</details>

Community and support
=====================

Kurtosis is a free and source-available product maintained by the [Kurtosis][kurtosis-tech] team. We'd love to hear from you and help where we can. You can engage with our team and our community in the following ways:

- Giving feedback via the `kurtosis feedback` command
- Filing an issue in our [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose)
- [Joining our Discord server][discord]
- [Following us on Twitter][twitter]
- Emailing us via the CLI: `kurtosis feedback --email`
- [Hop on a call to chat with us](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding)

<!-------- ONLY LINKS BELOW THIS POINT -------->
[enclave]: https://docs.kurtosis.com/explanations/architecture#enclaves
[awesome-kurtosis]: https://github.com/kurtosis-tech/awesome-kurtosis#readme
[quickstart-reference]: https://docs.kurtosis.com/quickstart
[discord]: https://discord.gg/Es7QHbY4
[kurtosis-tech]: https://github.com/kurtosis-tech
[docs]: https://docs.kurtosis.com
[twitter]: https://twitter.com/KurtosisTech
[starlark-explanation]: https://docs.kurtosis.com/explanations/starlark


<img src="./logo.png" width="1200">

----
What is Kurtosis?
=================
[Kurtosis](https://www.kurtosis.com) is a composable build system for multi-container test environments. Kurtosis makes it easier for developers to set up test environments that require dynamic setup logic (e.g. passing IPs or runtime-generated data between services) or programmatic data seeding. 

To read more about "why Kurtosis?", go [here](https://docs.kurtosis.com/explanations/what-is-kurtosis).

To read about the architecture, go [here](https://docs.kurtosis.com/explanations/architecture).


Running Kurtosis
================

### Install

Follow the instructions [here](https://docs.kurtosis.com/install).

### Run
Kurtosis can be used to create ephemeral environments called [enclaves][enclave]. We'll create a simple [Starlark][starlark-explanation] script to specify what we want our enclave to look like and what it will contain. Let's write one that spins up multiple replicas of `httpd`:

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

Running the following will give us an enclave with three services:

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
To see a more in-depth example and explanation of Kurtosis' capabilities, visit our [quickstart][quickstart-reference].

### More examples

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
- [Emailing us](mailto:feedback@kurtosistech.com) (also available as `kurtosis feedback --email`)
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

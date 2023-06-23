---
title: Kurtosis Architecture
sidebar_label: Architecture
sidebar_position: 2
---

![Kurtosis Architecture](/img/explanations/kurtosis-architecture.png)

Kurtosis At A Macro Level
-------------------------
At a macro level, Kurtosis is a set of containers, deployed on top of a container orchestrator (e.g. Docker, Kubernetes), that expose APIs. All interaction with Kurtosis is done via APIs. After Kurtosis receives a request, it usually reads or modifies some state in the container orchestrator. Kurtosis therefore serves as an abstraction layer atop the container orchestrator.

One Layer Deeper
----------------
To understand what Kurtosis itself does, we'll need to understand environments. Kurtosis' philosophy is that [the distributed nature of modern software means that modern software development now happens at the environment level][why-we-built-kurtosis]. To respond to this need, environments are a first-class concept in Kurtosis: easy to create, easy to inspect, easy to modify, and easy to destroy.

Therefore, the job of Kurtosis is to receive requests from the user and translate them to instructions for the underlying container orchestration engine. These requests can be simple commands that map one-to-one to instructions to the underlying container orchestrator (e.g. "add service X to environment Y"), or they can be Kurtosis-only commands that require complex interaction with the container orchestrator (e.g. "create a simulated network partition in environment X").

Enclaves
--------
Kurtosis implements environments as a first-class concept using [enclaves][enclaves-reference]. An enclave can be thought of as an "environment container" - an isolated place for a user to run an environment that is easy to create, manage, and destroy. Each enclave is separated from the other enclaves: no network communication can happen between them. Enclaves are also cheap: Kurtosis can manage arbitrary numbers of enclaves, limited only by the underlying hardware.

Example: Some enclaves running in, as displayed by [the Kurtosis CLI][cli-reference]:

```
UUID           Name      Status     Creation Time
a72b68e510fe   test      RUNNING    Thu, 30 Mar 2023 09:12:17 -03
9e8c913754bf   local     RUNNING    Thu, 30 Mar 2023 09:13:04 -03
```

Engine Container
----------------
The first type of container that Kurtosis creates is the Kurtosis engine container. This container's API is principally responsible for managing enclaves. This includes [creating enclaves][enclave-add-reference], [listing enclaves][enclave-ls-reference], [inspecting enclaves][enclave-inspect-reference], and [removing enclaves][enclave-rm-reference].

Example: A Kurtosis engine container running in Docker:

```console
bfb5627ff511   kurtosistech/engine:0.70.7                        "/bin/sh -c ./kurtos…"   10 hours ago   Up 10 hours   0.0.0.0:9710-9711->9710-9711/tcp                   kurtosis-engine--f84ce1f4c5ea410080e774cfea0ea0a4
```

Services
--------
Enclaves contain distributed applications, and distributed applications are composed of services. In Kurtosis, a service is a container that exposes ports. Services may also depend on other services (e.g. an API server depending on a database). Each enclave can have an arbitrary numbers of services, limited only by the underlying hardware.

Example: A pair of Nginx services running inside an enclave called `test`, as reported by the Kurtosis CLI:

```
UUID:                         2e42f9fd7b854eabb04f71a15bd1b55f
Name:                         test
Status:                       RUNNING
Creation Time:                        Thu, 24 Nov 2022 11:11:34 -03

========================================== User Services ==========================================
GUID                ID       Ports                             Status
nginx-1669299161    nginx    http: 80/tcp -> 127.0.0.1:60785   RUNNING
nginx2-1669299176   nginx2   http: 80/tcp -> 127.0.0.1:60794   RUNNING
```

API Container
-------------
The second type of container that Kurtosis creates is the API container. One API container is created per enclave, and each API container is responsible for managing interactions with its own enclave. All manipulation of an enclave's contents happens via API calls to the enclave's API container. 

Example: a Kurtosis API container running in Docker:

```
3c0b6ab7bb85   kurtosistech/core:0.70.7                          "/bin/sh -c ./api-co…"   20 hours ago    Up 20 hours    0.0.0.0:58419->7443/tcp, 0.0.0.0:58418->7444/tcp   kurtosis-api--6babc3090ad04184b2094901a7ead7b4
```

Starlark
--------
Distributed system definitions are complex. Therefore, there are many, many ways to instantiate, configure, and manipulate an enclave. To provide the required power, manipulations to an enclave are expressed using [Starlark][starlark-reference] scripts.

To manipulate an enclave, users upload Starlark scripts to the API container. The API container executes the instructions in the script, and the enclave's contents will be mutated.

Example: This Starlark snippet from [the quickstart][quickstart] launches a Postgres container:

```python
def run(plan, args):
    postgres = plan.add_service(
        service_name = "postgres",
        config = ServiceConfig(
            image = "postgres:15.2-alpine",
            ports = {
                "postgresql": PortSpec(5432, application_protocol = "postgresql"),
            },
            env_vars = {
                "POSTGRES_PASSWORD": "password",
            },
        ),
    )
```

For a list of all the Kurtosis Starlark instructions, [see here][starlark-code-reference].

SDK
----------------
All interactions with Kurtosis happen through API requests to the Kurtosis containers. To assist with calling the API, [we provide an SDK in various languages](https://github.com/kurtosis-tech/kurtosis/tree/main/api). Anything the Kurtosis can do will be available via the API and, therefore, via the SDK.

To see documentation for our SDK, [go here][SDK-reference].

For day-to-day operation, we also provide [a CLI][cli-reference]. This is simply a Go CLI wrapped around the Go Kurtosis SDK.

<!-------------- ONLY LINKS BELOW HERE --------------------->
[cli-reference]: ../cli-reference/index.md
[reusable-environment-definitions]: ./reusable-environment-definitions.md
[why-we-built-kurtosis]: ./why-we-built-kurtosis.md
[starlark-reference]: ../concepts-reference/starlark.md
[starlark-code-reference]: ../starlark-reference/index.md
[enclaves-reference]: ../concepts-reference/enclaves.md
[enclave-add-reference]: ../cli-reference/enclave-add.md
[enclave-ls-reference]: ../cli-reference/enclave-ls.md
[enclave-inspect-reference]: ../cli-reference/enclave-inspect.md
[enclave-rm-reference]: ../cli-reference/enclave-rm.md
[quickstart]: ../quickstart.md
[sdk-reference]: ../sdk.md

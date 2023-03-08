---
title: Kurtosis Architecture
sidebar_label: Architecture
sidebar_position: 2
---

![Kurtosis Architecture](@site/static/img/explanations/kurtosis-architecture.png)

Kurtosis At A Macro Level
-------------------------
At a macro level, Kurtosis is an engine (a set of Kurtosis servers) deployed on top of a container orchestrator (e.g. Docker, Kubernetes). All interaction with Kurtosis is done via the Kurtosis APIs. After the Kurtosis engine receives a request, it usually modifies some state inside the container orchestrator. Kurtosis therefore serves as an abstraction layer atop the container orchestrator, so that code written for Kurtosis is orchestrator-agnostic.

To understand what the Kurtosis engine does, we'll need to understand environments. Kurtosis' philosophy is that [the distributed nature of modern software means that modern software development now happens at the environment level][what-is-kurtosis]. To respond to this need, environments are a first-class concept in Kurtosis: easy to create, easy to inspect, easy to modify, and easy to destroy.

Therefore, the job of the Kurtosis engine is to receive requests from the client and translate them to instructions for the underlying container orchestration engine. These requests can be simple commands that map 1:1 to instructions to the underlying container orchestrator (e.g. "add service X to environment Y"), or they can be Kurtosis-only commands that require complex interaction with the container orchestrator (e.g. "divide environment X in two with a simulated network partition").

Enclaves
--------
Kurtosis implements "environments as a first-class" concept using **enclaves**. An enclave can be thought of an "environment container" - an isolated place for a user to run an environment that is easy to create, manage, and destroy. Each enclave is separate from the other enclaves: no network communication can happen between them. Enclaves are also cheap: each Kurtosis engine can manage arbitrary numbers of enclaves, limited only by the underlying hardware.

Example: Some enclaves running in a Kurtosis engine, as displayed by [the Kurtosis CLI][installation]:

```
EnclaveUUID                         Name     Status   Creation Time
a525cee593af4b45aa15785e87d3b7c9    local    RUNNING   Thu, 24 Nov 2022 14:11:27 UTC
edf36be917504e449a1648cf8d6c78a4    test     RUNNING   Thu, 24 Nov 2022 14:11:34 UTC
```

Services
--------
Enclaves contain distributed applications, and distributed applications have services. A service in Kurtosis is a container that exposes ports, and services may depend on other services (e.g. an API server depending on a database). Each enclave can have an arbitrary numbers of services, limited only by the underlying hardware.

Example: A pair of Nginx services running inside an enclave called `test`, as reported by the Kurtosis CLI:

```
Enclave UUID:                         2e42f9fd7b854eabb04f71a15bd1b55f
Enclave Name:                         test
Enclave Status:                       RUNNING
Creation Time:                        Thu, 24 Nov 2022 11:11:34 -03
API Container Status:                 RUNNING
API Container Host GRPC Port:         127.0.0.1:60768
API Container Host GRPC Proxy Port:   127.0.0.1:60769

========================================== User Services ==========================================
GUID                ID       Ports                             Status
nginx-1669299161    nginx    http: 80/tcp -> 127.0.0.1:60785   RUNNING
nginx2-1669299176   nginx2   http: 80/tcp -> 127.0.0.1:60794   RUNNING
```

SDKs
----
All interaction with Kurtosis happens via API requests to the Kurtosis engine. To assist with calling the API, [we provide SDKs in various languages](https://github.com/kurtosis-tech/kurtosis-sdk). Anything the Kurtosis engine is capable of doing will be available via the API and, therefore, via the SDKs.

For day-to-day operation, we also provide [a CLI ][installation] (usage guide [here][cli-usage]). This is a Go CLI that uses the Go SDK underneath.

Kurtosis Instruction Language
-----------------------------
Distributed system definitions are complex. To allow users to express their system in the simplest way possible while still fulfilling the required [properties of a reusable environment definition][reusable-environment-definitions], the Kurtosis engine provides users with the ability [to define and manipulate enclaves using Google's Starlark configuration language][starlark-explanation]. The Kurtosis engine contains a Starlark interpreter, and users can [send Starlark instructions][starlark-instructions] via the Kurtosis SDK to tell the engine what to do with an enclave. This allows users to define their environments as code.

For a reference list of the available Starlark instructions, [see here][starlark-instructions].

<!-------------- ONLY LINKS BELOW HERE --------------------->
[installation]: ../guides/installing-the-cli.md
[cli-usage]: ../reference/cli/cli.md
[reusable-environment-definitions]: ./reusable-environment-definitions.md
[what-is-kurtosis]: ./what-is-kurtosis.md
[starlark-explanation]: ./starlark.md
[starlark-instructions]: ../reference/starlark-instructions.md
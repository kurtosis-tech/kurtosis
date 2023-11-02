---
title: Basic Concepts
sidebar_label: Basic Concepts
slug: '/basic-concepts'
sidebar_position: 3
---

Package
-----------

A package defines the set up logic for a containerized backend. Packages can be accessed via Github locators.

```bash 
kurtosis run github.com/kurtosis-tech/basic-service-package
```

Packages can accept a set of parameters, defined by the package author, which enable the package consumer to modify their deployed backend at a high-level without needing to know how to configure each individual service.

```bash
kurtosis run github.com/kurtosis-tech/basic-service-package \
  '{"service_a_count": 2, 
    "service_b_count": 2, 
    "service_c_count": 1,
    "party_mode": true}'
```

Enclave
-----------

An enclave is a "walled garden" in which Kurtosis runs a containerized backend. An enclave contains all of the containers, subnets, files, and log aggregation tooling relevant for the environment. The main purposes of an enclave are resource isolation and "garbage collection". With an "enclave remove" command (`kurtosis enclave rm`), the end user can destroy all of the resources used to set up their environment and leave nothing hanging around on their machine(s).

Plan
-----------

The "plan" is the series of instructions, encoded in a package, that runs in an enclave. Each instruction in a plan is a basic building block for spinning up a containerized backend. You can see the plan that a package will run by "dry-running" the package:

```bash
kurtosis run --dry-run github.com/kurtosis-tech/basic-service-package
```

<details><summary><b>Output</b></summary>

```title="Steps in the Plan"
> render_templates

> add_services configs={"service-a-1": ServiceConfig(image="h4ck3rk3y/service-a", ports={"frontend": PortSpec(number=8501, application_protocol="http")}, files={"/app/config": "slender-boulder"})}

> render_templates

> add_services configs={"service-b-1": ServiceConfig(image="h4ck3rk3y/service-b", ports={"frontend": PortSpec(number=8501, application_protocol="http")}, files={"/app/config": "purple-comet"}, cmd=["false"])}

> render_templates

> add_services configs={"service-c-1": ServiceConfig(image="h4ck3rk3y/service-c", ports={"frontend": PortSpec(number=8501, application_protocol="http")}, files={"/app/config": "arctic-oak"}, env_vars={"PARTY_MODE": "false"})}
```

</details>

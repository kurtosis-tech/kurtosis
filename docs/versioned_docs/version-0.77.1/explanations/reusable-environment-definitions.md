---
title: Reusable Environment Definitions
sidebar_label: Reusable Environment Definitions
---

Why are reusable environment definitions hard?
----------------------------------------------
We have many tools for defining and modifying environments: Bash/Python scripts, Ansible, Docker Compose, Helm, and Terraform. Yet, none have proven successful at reuse across [the plethora of environments in today's world][why-we-built-kurtosis]. To see why, we'll focus on the three most common environment types: local Dev, ephemeral Test in CI, and Prod.

Environment definitions in Dev, Test, and Prod share some common requirements:

- They must be easy to read and write
- They must be parameterizable
- They must handle data (e.g. on-disk files, database dumps, etc.)
- They should do some amount of validation

They also have distinct differences:

- In Dev, environment definitions must be loose and easy to modify, are often not checked into source control, and are rarely shared due to the prototyping nature. However, developers _do_ want to leverage other public environment definitions to form their local development environments quickly (e.g. to combine public definitions for Postgres and Elasticsearch to form a Postgres+Elasticsearch local development environment).
- In Test, environment definitions should be source-controlled, but are rarely shared given how specific the environment definition is to the scenario being tested. The order in which events happen is very important, so determinism and ordering of events matter.
- In Prod, environment definitions must be source-controlled. They may be shared, but can only be modified by authorized individuals. They are nearly always declarative: idempotence is very important in Prod, along with the ability to make the minimal set of changes (e.g. add just a few new services without needing to restart the entire stack).

Why aren't the current solutions reusable?
------------------------------------------
With this lens, we see why none of the most common solutions can be reused across Dev, Test, and Prod:

- Scripting can be used for Dev and Test, but instantiating just a part of the environment requires extensive parameterization. Scripting is also not declarative, and requires the author to build in idempotence. Validation and data-handling is left to the author. Finally, scripts can get very complex and require specialized knowledge to understand.
- Ansible behaves like scripting with idempotence and validation on top. It is better in the Prod usecase than scripting, but doesn't ease the Dev/Test usecase of instantiating just a part of the system.
- Docker Compose is great for Dev and even for some Test, but fails in Prod for its lack of idempotence and its requirement to bring the entire stack down and up each time. It has no validation, little parameterizability, and Docker Compose files cannot be plugged together.
- Helm is excellent for the Prod usecase for its idempotence, parameterizability, and emphasis on sharing, but Helm charts are complex and difficult to compose or decompose. Like Docker Compose, data-handling is via volumes only. The only execution mode is declarative, so Helm only fills the Test usecase when mixed with a procedural language.
- Terraform, like Helm, hits the Prod usecase very well. However, like Helm, Terraform can only be executed in declarative mode; Test usecases with Terraform therefore need a procedural langauge to sequence events.

What does a reusable solution look like?
----------------------------------------
Kurtosis believes that any environment definition that aims to be reusable across Dev, Test, and Prod must have six properties:

1. **Composability:** The user should be able to combine two or more environment definitions to form a new one (e.g. Postgres + Elasticsearch).
    - In Dev, Test, and Prod, this allows modularization of definitions
1. **Decomposability:** The user should be able to take an existing environment definition and strip out the parts they're not interested in to form a smaller environment definition (e.g. take the large Prod environment definition and instantiate only a small portion of it).
    - In Dev, Test, and Prod, this consumers of third-party definitions can select only the sub-components of the definition that are of interest
    - In Dev, developers can select just the parts of their Prod system that they're working on
1. **Safety:** The user should be able to know whether the environment definition will work before instantiating it (analogous to type-checking - e.g. verifying all the ports match up, all the IP addresses match up, all the container images are available, etc.).
    - In Dev, Test, and Prod, this left-shifts classes of errors from runtime to validation time resulting in a tighter feedback loop
1. **Parameterizability:** An environment definition should be able to accept parameters (e.g. define the desired number of Elasticsearch nodes).
    - In Dev, Test, and Prod, parameterizability is essential to keeping environment definitions DRY
1. **Pluggability of Data:** The data used across Dev, Test, and Prod varies so widely that the user should be able to configure which data to use.
    - In Dev, Test, and Prod, this allows for a definition to be reusable even when the data isn't the same
1. **Portability:** An environment definition author should be able to share their work and be confident that it can be consumed.
    - In Dev, Test, and Prod, this allows for reuse of definitions
    - In Test, test cases that failed on a CI machine can be reproduced on a developer's machine

Kurtosis environment definitions (written in [Starlark][starlark-reference]) are designed with these six properties in mind.


<!--------------------- ONLY LINKS BELOW HERE ------------------------->
[starlark-reference]: ../concepts-reference/starlark.md
[why-we-built-kurtosis]: ./why-we-built-kurtosis.md

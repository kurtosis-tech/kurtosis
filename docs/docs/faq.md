---
title: FAQ
sidebar_label: FAQ
slug: /faq
---

Why can't I do X in Starlark?
-----------------------------
Starlark is intended to be a configuration and orchestration language, not a general-purpose programming language. It is excellent at simplicity, readability, and determinism, and terrible at general-purpose programming. We want to use Starlark for what it's good at, while making it easy for you to call down to whatever general-purpose programming you need for more complex logic.

Therefore, Kurtosis provides:

- [`plan.run_sh`](./starlark-reference/plan.md#run_sh) for running Bash tasks on a disposable container
- [`plan.run_python`](./starlark-reference/plan.md#run_python) for running Python tasks on a disposable container
- [`plan.exec`](./starlark-reference/plan.md#exec) for running Bash on a service

All of these let you customize the image to run on, so you can functionally call any code in any language using Kurtosis.

What is Kurtosis building next?
-------------------------------
Great question, check out our [roadmap page](./roadmap.md) for the latest details on where we plan to take Kurtosis next.

Why am I getting rate limited by Dockerhub when pulling images?
---------------------------------------------------------------
Kurtosis will first try to use your locally cached container images before pulling any image from Dockerhub. If you are getting rate limited by Dockerhub when pulling images, it likely means you have exceeded the [limits set by Docker](https://docs.docker.com/docker-hub/download-rate-limit/). 

Does Kurtosis support other container registries or libraries?
--------------------------------------------------------------
Currently, Kurtosis only supports Dockerhub. If your project or team requires a different type of container registry, please let us know by [filing an issue in our Github](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=feature+request&projects=&template=feature-request.yml) or letting us know in [Discord](https://discord.gg/jJFG7XBqcY). 

Does Kurtosis pull a container image down each time I run a package?
--------------------------------------------------------------------
Kurtosis will always first check the local cache for a given container image for each `kurtosis run` before pulling the image from an external registry (e.g. Dockerhub).

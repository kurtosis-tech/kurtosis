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

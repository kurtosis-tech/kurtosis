---
title: Best Practices
sidebar_label: Best Practices
slug: /best-practices
---

Passing package arguments to the CLI
-------------------------------
Passing [package arguments][args-concepts-reference] to the CLI can get hairy due to the interaction between Bash and JSON quotes. The following are tips to make your life easier:

1. **When you have a small number of arguments:** surround the arguments with single quotes so you don't have to escape double quotes in your JSON. E.g.:
   ```bash
   kurtosis run github.com/user/repo '{"some_param":5,"some_other_param":"My value"}'
   ```
1. **When you have a large number of arguments:** put them in a `.json` file and use [Bash command substitution](https://www.gnu.org/software/bash/manual/html_node/Command-Substitution.html) _inside double quotes_ to slot them into the `kurtosis run` command. E.g.:
   ```bash
   kurtosis run github.com/user/repo "$(cat my-args.json)"
   ```
   The double quotes around the `$(cat my-args.json)` are important so any spaces inside `my-args.json` don't fool Bash into thinking you're passing in two separate arguments.

Choosing the right wait
-----------------------
Kurtosis has three different types of waits. Described here are the three, with tips on when to use each:

1. Automatic waiting on port availability when a service starts (enabled by default; can be configured with [`PortSpec.wait`][port-spec-starlark-reference])
    - Should be sufficient for most usecases
    - Requires little-to-no extra configuration
    - Will cause parallel `Plan.add_services` to fail, allowing for quick aborting
1. Waiting on [`ReadyCondition`][ready-condition-starlark-reference]s (configured in [`ServiceConfig`][service-config-starlark-reference])
    - Allows for more advanced checking (e.g. require a certain HTTP response body, ensure a CLI call returns success, etc.)
    - More complex to configure
    - Will cause parallel `Plan.add_services` to fail, allowing for quick aborting
1. The [`Plan.wait`][plan-wait-starlark-reference]
    - Most useful for asserting the system has reached a desired state in tests (e.g. wait until data shows up after loading)
    - More complex to configure
    - Cannot be used to short-circuit `Plan.add_services`

<!---------------------------------------- ONLY LINKS BELOW HERE!!! ----------------------------------->
[args-concepts-reference]: ./concepts-reference/args.md

[service-config-starlark-reference]: ./starlark-reference/service-config.md
[port-spec-starlark-reference]: ./starlark-reference/port-spec.md
[ready-condition-starlark-reference]: ./starlark-reference/ready-condition.md
[plan-wait-starlark-reference]: ./starlark-reference/plan.md#wait

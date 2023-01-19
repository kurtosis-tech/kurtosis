---
title: CLI
sidebar_label: CLI
slug: /cli
sidebar_position: 1
---


The [Kurtosis CLI](../guides/installing-the-cli.md) is the main way to interact with Kurtosis. This document will present some common CLI workflows.

:::tip
The `kurtosis` command, and all of its subcommands, will print helptext when passed the `-h` flag. You can use this at any time to see information on the command you're trying to run. For example:
```
kurtosis service -h
```
:::

:::tip
Kurtosis supports tab-completion, and we strongly recommend [installing it][adding-tab-completion] for the best experience!
:::

### Initialize configuration
When the Kurtosis CLI is executed for the first time on a machine, we ask you to make a choice about whether [you'd like to send anonymized usage metrics to help us make the product better](../explanations/metrics-philosophy.md). To make this election non-interactively, you can run either:

```bash
kurtosis config init send-metrics
```

to send anonymized metrics to improve the product or

```bash
kurtosis config init dont-send-metrics
```

if you'd prefer not to.

### Get the CLI version
The CLI version can be printed with the following:

```
kurtosis version
```

### Restart the engine
The CLI interacts with the Kurtosis engine, which is a very lightweight container. The CLI will start the engine container automatically for you and you should never need to start it manually, but you might need to restart the engine after a CLI upgrade. To do so, run:

```bash
kurtosis engine restart
```

### Check engine status
The engine's version and status can be printed with:

```bash
kurtosis engine status
```

### Stop the engine
To stop the engine, run:

```bash
kurtosis engine stop
```

### Run Starlark
A single Starlark script can be ran with:

```bash
kurtosis run script.star
```

Adding the `--dry-run` flag will print the changes without executing them.

A [Kurtosis package][packages-reference] on your local machine can be run with:

```bash
kurtosis run /path/to/package/on/your/machine
```

A [runnable Kurtosis package][packages-reference] published to GitHub can be run like so:

```bash
kurtosis run github.com/package-author/package-repo
```

Arguments can be provided to a Kurtosis package (either local or from GitHub) by passing a JSON-serialized object with args argument, which is the second positional argument you pass to `kurtosis run` like:

```bash
# Local package
kurtosis run /path/to/package/on/your/machine '{"company":"Kurtosis"}'

# GitHub package
kurtosis run github.com/package-author/package-repo '{"company":"Kurtosis"}'
```

### Create an enclave
Your distributed applications run in [enclaves][enclaves-explanation]. They are isolated from each other, to ensure they don't interfere with each other. To create a new, empty enclave, run:

```bash
kurtosis enclave add
```

### List enclaves
To see all the enclaves in Kurtosis, run:

```bash
kurtosis enclave ls
```

The enclave UUIDs and names that are printed will be used in enclave manipulation commands.

### View enclave details
To view detailed information about a given enclave, run:

```bash
kurtosis enclave inspect $THE_ENCLAVE_IDENTIFIER
```

This will print detailed information about:

- The enclave's status (running or stopped)
- The services inside the enclave (if any), and the information for accessing those services' ports from your local machine

### Dump enclave information to disk
You'll likely need to store enclave logs to disk at some point. You may want to have a log package if your CI fails, or you want to send debugging information to [the author of a Kurtosis package][packages]. Whatever the case may be, you can run:

```bash
kurtosis enclave dump $THE_ENCLAVE_IDENTIFIER $OUTPUT_DIRECTORY
```

You'll get the container logs & configuration in the output directory for further analysis & sharing.

### Delete an enclave
To delete an enclave and everything inside of it, run:

```bash
kurtosis enclave rm $THE_ENCLAVE_IDENTIFIER
```

Note that this will only delete stopped enclaves. To delete a running enclave, pass the `-f`/`--force` flag.

### Add a service to an enclave
To add a service to an enclave, run:

```bash
kurtosis service add $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $CONTAINER_IMAGE
```

Much like `docker run`, this command has multiple options available to customize the service that's started:

1. The `--entrypoint` flag can be passed in to override the binary the service runs
1. The `--env` flag can be used to specify a set of environment variables that should be set when running the service
1. The `--ports` flag can be used to set the ports that the service will listen on

To override the service's CMD, add a `--` after the image name and then pass in your CMD args like so:

```bash
kurtosis service add --entrypoint sh my-enclave test-service alpine -- -c "echo 'Hello world'"
```

### View a service's logs
To print the logs for a service, run:

```bash
kurtosis service logs $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER
```

The service identifier(name or uuid) is printed upon inspecting an enclave.

The following optional arguments can be used:
1. `-f`, `-follow` can be added to continue following the logs, similar to `tail -f`.
1. `--match=text` can be used for filtering the log lines containing the text.
1. `--regex-match="regex"` can be used for filtering the log lines containing the regex. This filter will also work for text but will have degraded performance.
1. `-v`, `--invert-match` can be used to invert the filter condition specified by either `--match` or `--regex-match`. Log lines NOT containing the match will be returned.

Important: `--match` and `--regex-match` flags cannot be used at the same time. You should either use one or the other.


### Run commands inside a service container
You might need to get access to a shell on a given service container. To do so, run:

```bash
kurtosis service shell $THE_ENCLAVE_ID $THE_SERVICE_IDENTIFIER
```

### Delete a service from an enclave
Services can be deleted from an enclave like so:

```bash
kurtosis service rm $THE_ENCLAVE_ID $THE_SERVICE_IDENTIFIER
```

**NOTE:** To avoid destroying debugging information, Kurtosis will leave removed services inside the Docker engine. They will be stopped and won't show up in the list of active services in the enclave, but you'll still be able to access them (e.g. using `service logs`) by their service GUID (available via `enclave inspect`).

### Upload files to an enclave
Files can be stored as a [files artifact][files-artifacts] inside an enclave by uploading them:

```bash
kurtosis files upload $PATH_TO_FILES
```

### Clean Kurtosis
Kurtosis defaults to leaving enclave artifacts (containers, volumes, etc.) around so that you can refer back them for debugging. To clean up artifacts from stopped enclaves, run:

```bash
kurtosis clean
```

To remove artifacts from _all_ enclaves (including running ones), add the `-a`/`--all` flag.

NOTE: This will not stop the Kurtosis engine itself! To do so, see "Stopping the engine" above.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[packages]: ../reference/packages.md
[enclaves-explanation]: ../explanations/architecture.md#enclaves
[adding-tab-completion]: ../guides/adding-tab-completion.md
[files-artifacts]: ./files-artifacts.md

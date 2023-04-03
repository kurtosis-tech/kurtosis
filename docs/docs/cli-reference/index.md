---
id: index
title: CLI Introduction
sidebar_label: Introduction
slug: /cli
sidebar_position: 1
---

The Kurtosis CLI is a Go CLI wrapped around the Kurtosis Go [client library][client-library-reference]. This section will go through the most common Kurtosis CLI commands and some useful tips on getting started. If you have not already done so, the CLI can be installed by following the instructions [here][installing-the-cli].


:::tip
Kurtosis supports command-line completion; we recommend [installing it][adding-command-line-completion] for the best experience.
:::

### Configuration file path
To locate where the Kurtosis configuration file is on your machine, simply use:

```bash
kurtosis config path
```
to print out the file path of the `kurtosis-config.yml` file.

### Get the CLI version
The version of the CLI and the currently-running engine can be printed with the following:

```
kurtosis version
```

### Global Flags
Kurtosis CLI supports two global flags - `help` and `cli-log-level`. These flags can be used with any Kurtosis CLI commands.

#### -h or --help
This flag prints the help text for all commands and subcommands. You can use this at any time to see information on the command you're trying to run. For example:
```
kurtosis service -h
```

#### cli-log-level
This flag sets the level of details that the Kurtosis CLI will print logs with - by default it only logs `info` level logs to the CLI. The following other log levels are supported by Kurtosis -
```panic|fatal|error|warning|info|debug|trace```. For example, logs with error level can be printed using the command below:- 

```
kurtosis run --cli-level-log error github.com/package-author/package-repo 
```

:::info

Users can use the `--cli-log-level` flag to display the entire stack trace to the CLI. By default the entire stack trace is saved to the `kurtosis-cli.log` file. The command below, for example, will display the entire stack-traces to the CLI for debugging purposes:

```
kurtosis run --cli-log-level debug github.com/package-author/package-repo 
```
:::




<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[adding-command-line-completion]: ../guides/adding-command-line-completion.md
[installing-the-cli]: ../guides/installing-the-cli.md
[client-library-reference]: ../client-libs-reference.md

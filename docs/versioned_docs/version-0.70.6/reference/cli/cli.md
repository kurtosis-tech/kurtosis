---
title: Getting Started
sidebar_label: CLI
slug: /cli
sidebar_position: 1
---

This section will go through the most common Kurtosis CLI commands and some useful tips on getting started. If you have not already done so, the CLI can be installed by following the instructions [here][installing-the-cli].

:::tip
The `kurtosis` command, and all of its subcommands, will print helptext when passed the `-h` or `--help` flag. You can use this at any time to see information on the command you're trying to run. For example:
```
kurtosis service -h
```
:::

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
The CLI version along with the currently running engine if any; can be printed with the following:

```
kurtosis version
```

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[adding-command-line-completion]: ../../guides/adding-command-line-completion.md
[installing-the-cli]: ../../guides/installing-the-cli.md

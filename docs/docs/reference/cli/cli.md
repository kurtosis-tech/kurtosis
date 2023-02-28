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
Kurtosis supports tab-completion, and we strongly recommend [installing it][adding-tab-completion] for the best experience!
:::

### Toggle Analytics
On installation Kurtosis enables anonymized [analytics][metrics-philosophy-reference] by default. You can toggle this functionality simply by running:
We identify every user's machines with a hash value of the `unique machine id of most host OS's`, which we call the `analytics ID`, and this allow us to analyze the user experience flow.

```bash
kurtosis analytics enable
```

to enable the sending of anonymized metrics to improve the product, or:

```bash
kurtosis analytics disable
```

if you would prefer not to.

You may optionally pass in the following flags with this command:
* `--id`: Prints the `analytics ID` 

### Configuration file path
To locate where the Kurtosis configuration file is on your machine, simply use:

```bash
kurtosis config path
```
to print out the file path of the `kurtosis-config.yml` file.

### Get the CLI version
The CLI version can be printed with the following:

```
kurtosis version
```

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[adding-tab-completion]: ../../guides/adding-tab-completion.md
[installing-the-cli]: ../../guides/installing-the-cli.md
[metrics-philosophy-reference]: ../../explanations/metrics-philosophy.md

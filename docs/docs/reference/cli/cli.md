---
title: CLI
sidebar_label: CLI
slug: /cli
sidebar_position: 1
---

The [Kurtosis CLI][installing-the-cli] is the main way to interact with Kurtosis. This document will present some common CLI workflows.

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
When the Kurtosis CLI is executed for the first time on a machine, we ask you to make a choice about whether [you'd like to send anonymized usage metrics to help us make the product better][metrics-philosophy-reference]. To make this election non-interactively, you can run either:

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

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[adding-tab-completion]: ../../guides/adding-tab-completion.md
[installing-the-cli]: ../../guides/installing-the-cli.md
[metrics-philosophy-reference]: ../../explanations/metrics-philosophy.md

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

### Toggle Telemetry
On installation Kurtosis enables anonymized [telemetry](../explanations/metrics-philosophy.md) by default. You can toggle it,

```bash
kurtosis analytics enable
```

to send anonymized metrics to improve the product or

```bash
kurtosis analytics disable
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

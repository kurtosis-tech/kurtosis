---
title: config path
sidebar_label: config path
slug: /config-path
---

The `kurtosis config path` command displays the path to the Kurtosis CLI config YAML file. This file is used to configure Kurtosis CLI behaviour.

To see the full set of configuration values available:

1. Open [the directory containing the versions of Kurtosis config](https://github.com/kurtosis-tech/kurtosis/tree/main/cli/cli/kurtosis_config/overrides_objects)
2. Select the most recent (highest) version
3. Explore the various config objects inside, starting with the `kurtosis_config_vX.go` top-level object (each `struct` represents a YAML object inside the configuration)

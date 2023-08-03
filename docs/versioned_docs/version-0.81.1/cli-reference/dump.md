---
title: dump
sidebar_label: dump
slug: /dump
---

You might need to store the entire state of Kurtosis to disk at some point. You may want to have a log package if your CI fails, or you want to send debugging information to [the author of a Kurtosis package][packages-reference]. Whatever the case may be, you can run:

```bash
kurtosis dump $OUTPUT_DIRECTORY
```
You will get the container logs & configuration in the output directory for further analysis & sharing. This would contain all engines & enclaves.

If you don't specify the `$OUTPUT_DIRECTORY` Kurtosis will dump it to a directory with a name following the schema `kurtosis-dump--TIMESTAMP`.

<!-------------------- ONLY LINKS BELOW THIS POINT ----------------------->
[packages-reference]: ../concepts-reference/packages.md

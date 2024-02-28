---
title: lint
sidebar_label: lint
slug: /lint
---

The following command can be used to lint Starlark files in the given package

To get running quickly, simply run

```bash
kurtosis lint .
```

This will lint all the Starlark files in the given package

Instead of just finding linting issues if you want to format the files as well use the `--format` flag

```bash
kurtosis lint . --format
```

You can also lint a specific file via

```bash
kurtosis lint main.star
```

Or to lint multiple files or directories at the same time

```bash
kurtosis lint this.star that.star also-this.star my-favorite-directory/
```

To validate a `main.star` doc string use the `-c` or the `--check-docstring` flag. Note that this requires
you pass a single path to a `main.star` or a single directory containing a `main.star`

```bash
kurtosis lint . -c
```
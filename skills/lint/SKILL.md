---
name: lint
description: Lint and format Kurtosis Starlark files. Check syntax, validate docstrings, and auto-format .star files. Use when writing or reviewing Starlark packages to ensure code quality.
compatibility: Requires kurtosis CLI.
metadata:
  author: ethpandaops
  version: "1.0"
---

# Lint

Lint and format Kurtosis Starlark (.star) files.

## Check formatting

```bash
# Check a single file
kurtosis lint main.star

# Check a directory
kurtosis lint ./my-package/

# Check multiple files
kurtosis lint main.star lib.star helpers.star
```

Returns non-zero exit code if formatting issues are found.

## Auto-format

Fix formatting in place:

```bash
kurtosis lint -f main.star

# Format all files in a package
kurtosis lint -f ./my-package/
```

## Check docstrings

Validate that the main function has a proper docstring:

```bash
kurtosis lint -c ./my-package/main.star

# Or point to the package directory
kurtosis lint -c ./my-package/
```

This checks that the `run` function has a valid docstring describing its parameters.

## CI integration

```bash
# Check formatting (fails if not formatted)
kurtosis lint ./my-package/

# Check docstrings
kurtosis lint -c ./my-package/

# Auto-format before commit
kurtosis lint -f ./my-package/
```

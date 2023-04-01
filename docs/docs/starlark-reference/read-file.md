---
title: read_file
sidebar_label: read_file
---

The `read_file` function reads the contents of a file specified by the given [locator][locators-reference], and requires that the Starlark script is part of a [package][packages-reference]. `read_file` executes [at interpretation time][multi-phase-runs-reference] so the file contents won't be displayed in the preview.

```python
contents = read_file(
    # The Kurtosis locator of the file to read.
    # MANDATORY
    src = "github.com/kurtosis-tech/datastore-army-package/README.md",
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[locators-reference]: ../concepts-reference/locators.md
[multi-phase-runs-reference]: ../concepts-reference/multi-phase-runs.md
[packages-reference]: ../concepts-reference/packages.md

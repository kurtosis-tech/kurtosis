---
title: read_file
sidebar_label: read_file
---

The `read_file` function reads the contents of a file specified by the given [locator][locators-reference] and executes [at interpretation time][multi-phase-runs-reference] so the file contents won't be displayed in the preview. This instruction returns the content of the file in a string type. Please note that the files being read from must themselves be part of a Kurtosis package, as explained [here](../concepts-reference/locators.md#important-package-restriction).

```python
read_file(
    # The Kurtosis locator of the file to read.
    # MANDATORY
    src = "LOCATOR",
)
```

For example:

```python
# Reading a file from a remote package using an absolute locator
remote_contents = read_file(
    src = "github.com/kurtosis-tech/datastore-army-package/README.md",
)

# Reading a file from inside the same package using a relative locator
local_contents = read_file(
    src = "./file.txt",
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[locators-reference]: ../concepts-reference/locators.md
[multi-phase-runs-reference]: ../concepts-reference/multi-phase-runs.md
[packages-reference]: ../concepts-reference/packages.md

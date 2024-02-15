---
title: import_module
sidebar_label: import_module
---

The `import_module` function imports the symbols from a Starlark script specified by the given [locator][locators-reference], and requires that the calling Starlark script is part of a [package][packages-reference].

```python
# Import remote code from another package using an absolute import
remote = import_module("github.com/foo/bar/src/lib.star")

# Simiarily, you can also import a specific version (e.g. 2.0) of an upstream package using an absolute import
remote_2 = import_module("github.com/foo/bar/src/lib.star@2.0")

# Import local code from the same package using a relative import
local = import_module("./local.star")

def run(plan):
    # Use code from the imported module
    remote.do_something(plan)

    local.do_something_else(plan)
```

NOTE: We chose not to use the normal Starlark `load` primitive due to its lack of namespacing. By default, the symbols imported by `load` are imported to the global namespace of the script that's importing them. We preferred module imports to be namespaced, in the same way that Python does by default with its `import` statement.


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[packages-reference]: ../../advanced-concepts/packages.md
[locators-reference]: ../../advanced-concepts/locators.md

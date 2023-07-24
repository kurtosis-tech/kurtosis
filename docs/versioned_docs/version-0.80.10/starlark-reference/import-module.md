---
title: import_module
sidebar_label: import_module
---


The `import_module` function imports the symbols from a Starlark script specified by the given [locator][locators-reference], and requires that the calling Starlark script is part of a [package][packages-reference].

```python
# Import the code to namespaced object
lib = import_module("github.com/foo/bar/src/lib.star")

# Use code from the imported module
lib.some_function()
lib.some_variable
```

NOTE: We chose not to use the normal Starlark `load` primitive due to its lack of namespacing. By default, the symbols imported by `load` are imported to the global namespace of the script that's importing them. We preferred module imports to be namespaced, in the same way that Python does by default with its `import` statement.



<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[packages-reference]: ../concepts-reference/packages.md
[locators-reference]: ../concepts-reference/locators.md

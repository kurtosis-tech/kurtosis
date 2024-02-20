---
title: StoreSpec
sidebar_label: StoreSpec
---

The `StoreSpec` is used to configure how to store a file on a [`run_sh`][run-sh-reference] or a [`run_python`][run-python-reference] container as a files artifact

```python
config = StoreSpec(
    # The path on the task container that needs to be stored in a files artifact
    # MANDATORY
    src = "/foo/bar",

    # The name of the files artifact; needs to be unique in the enclave
    # This is optional; if not provided Kurtosis will create a nature themed name
    # OPTIONAL
    name = "divine-pool"
)
```

Note the `StoreSpec` object is provided as a list to the `store` attribute of the `run_sh` and `run_python` instructions. Within
one list the `src` needs to be unique; while the `name` needs to be `unique` for the entire enclave.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[run-python-reference]: ./plan.md#run_python
[run-sh-reference]: ./plan.md#run_sh

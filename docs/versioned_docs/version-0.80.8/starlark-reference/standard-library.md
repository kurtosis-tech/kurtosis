---
title: Standard Library
sidebar_label: Standard Library
sidebar_position: 2
---

The following Starlark libraries are available in Kurtosis by default:

1. The Starlark [time](https://github.com/google/starlark-go/blob/master/lib/time/time.go#L18-L52) module (a collection of time-related functions)
2. The Starlark [json](https://github.com/google/starlark-go/blob/master/lib/json/json.go#L28-L74) module (allows `encode`, `decode` and `indent` JSON)
4. The Starlark [struct](https://github.com/google/starlark-go/blob/master/starlarkstruct/struct.go) builtin (allows you to create `structs` like the one used in [`add_service`][add-service-reference])

<!--------------------------------------- ONLY LINKS BELOW HERE -------------------------------->
[add-service-reference]: ./plan.md#add_services

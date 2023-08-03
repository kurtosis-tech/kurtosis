---
title: Subnetworks
sidebar_label: Subnetworks
---

:::tip

Want to get started with subnetworks? See how to easily simulate a networking failure [here][networking-failure-guide]

:::

A service started inside a Kurtosis enclave will by default be able to communicate with all other services inside the same enclave. To change that behavior, Kurtosis exposes the concept of *subnetworks*.

A subnetwork is a way to group together an arbitrary number of services from the same enclave. It is very lightweight and it can be instantiated by "tagging" each service with a subnetwork name (see [add_service][add-service] and [update_service][update-service]). A service added to an enclave without a subnetwork specified is added to the default subnetwork (called `default`).

All services with the same "subnetwork tag" will belong to the same subnetwork, and networking can then be configured _accross subnetworks_. All services in the same subnetwork (including the `default` one) will always be able to communicate.

The connection configuration between two subnetworks inherits a *default connection* defined in Kurtosis. It is set to allow all communications when the enclave starts but it can be changed using [set_connection][set-connection]. 

To have fine grained control, the connection between two specific subnetworks can be overridden also using the same [set_connection][set-connection]. The override can be removed using [remove_connection][remove-connection].

:::caution

This functionaility is only available for Kurtosis running on Docker. Kurtosis running on Kubernetes cannot use subnetworks yet, and instructions requiring it will throw an error.

:::

:::caution

This functionality must be enabled manually per enclave using the CLI. When running Starlark scripts or packages using this feature, add the `--with-subnetworks` optional flag.

:::


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->

[add-service]: ../starlark-reference/plan.md#add_service
[update-service]: ../starlark-reference/plan.md#update_service
[set-connection]: ../starlark-reference/plan.md#set_connection
[remove-connection]: ../starlark-reference/plan.md#remove_connection
[networking-failure-guide]: ../guides/simulating-networking-failure.md

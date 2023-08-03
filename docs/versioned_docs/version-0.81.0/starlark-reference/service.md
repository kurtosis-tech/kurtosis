---
title: Service
sidebar_label: Service
---

The `Service` object encapsulates service information returned by the [`Plan.add_service`][add-service-starlark-reference] and [`Plan.add_services`][add-services-starlark-reference] functions.

It has the following properties (all of which are [future references][future-references-concepts-reference], because [runtime values don't exist at interpretation time][multi-phase-runs-concepts-reference]):

```python
# The hostname of the service
service.hostname


# The IP address of the service
service.ip_address


# The name of the service
service.name


# A dictionary of port IDs -> PortSpec objects, as specified in the "ports" field 
# of the ServiceConfig used to create the service
# (see the PortSpec and ServiceConfig entries in the sidebar for more information)
service.ports

# For example:
some_port_spec = service.ports["some-port-id"]
```

Note that you cannot manually create a `Service` object; it is only returned by Kurtosis via `Plan.add_service` and `Plan.add_services`.


<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[add-service-starlark-reference]: ./plan.md#add_service
[add-services-starlark-reference]: ./plan.md#add_services

[future-references-concepts-reference]: ../concepts-reference/future-references.md
[multi-phase-runs-concepts-reference]: ../concepts-reference/multi-phase-runs.md

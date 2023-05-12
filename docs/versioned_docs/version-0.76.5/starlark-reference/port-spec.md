---
title: PortSpec
sidebar_label: PortSpec
---

The `PortSpec` constructor creates a `PortSpec` object that encapsulates information pertaining to a port.

```python
port_spec = PortSpec(
    # The port number which we want to expose
    # MANDATORY
    number = 3000,

    # Transport protocol for the port (can be either "TCP" or "UDP")
    # OPTIONAL (DEFAULT:"TCP")
    transport_protocol = "TCP",

    # Application protocol for the port that will be displayed in front of URLs containing the port
    # For example:
    #  "http" to get a URL of "http://..."
    #  "postgresql" to get a URL of "postgresql://..."
    # OPTIONAL (DEFAULT:"http")
    application_protocol = "http",
    
    # Kurtosis will automatically perform a check to ensure all declared UDP and TCP ports are open and ready for traffic and connections upon startup.
    # You may specify a custom wait timeout duration or disable the feature entirely.
    # You may specify a custom wait timeout duration with a string:
    #  wait = "2m"
    # Or, you can disable this feature by setting the value to None:
    #  wait = None
    # The feature is enabled by default with a default timeout of 15s
    # OPTIONAL (DEFAULT:"15s")
    wait = "4s"
)
```
The above constructor returns a `PortSpec` object that defines information about a port for use in [`add_service`](../concepts-reference/subnetworks.md).

The `wait` field represents the timeout duration that Kurtosis will use when checking whether or not a service's declared UDP or TCP port are open and ready for traffic and connections upon startup. This is the default way to perform a readiness check using Kurtosis. However, there are other ways to perform a readiness check. Specifically, you can also use [`ServiceConfig.ReadyConditions`][ready-conditions] to check if a service is ready with a a REST call, or you can use the [Wait][wait] instruction if you need to perform a one-off readiness check after service startup.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[future-references-reference]: ../concepts-reference/future-references.md
[add-service-reference]: ./plan.md#add_service
[ready-conditions]: ./ready-condition.md
[wait]: ./plan.md#wait

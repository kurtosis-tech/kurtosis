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
    
    # All the service's TCP and UDP ports are automatically check, Kurtosis check if the declared ports
    # are open before finishing the service startup.
    # We can set a custom wait time out value with or disable the feature or through this property
    # You can specify a custom wait time out using a Duration string:
    #  wait = "2m"
    # Or, you can disable this feature by setting the value in None:
    #  wait = None
    # The feature is enabled by default with a default timeout
    # OPTIONAL (DEFAULT:"15s")
    wait = "4s"
)
```
The above constructor returns a `PortSpec` object that defines information about a port for use in [`add_service`](../concepts-reference/subnetworks.md).

The wait field represents the timeout for the automatic ports opening check, which is the default way to check for readiness. There are others ways to check service readiness, you can use the [ServiceConfig.ReadyConditions][ready-conditions] which can be used to check if a service is ready through a REST call for instance, or you can use the [Wait][wait] instructions if you need a one-off check for readiness not attached to a service startup.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[future-references-reference]: ../concepts-reference/future-references.md
[add-service-reference]: ./plan.md#add_service
[ready-conditions]: ./ready-condition.md
[wait]: ./plan.md#wait

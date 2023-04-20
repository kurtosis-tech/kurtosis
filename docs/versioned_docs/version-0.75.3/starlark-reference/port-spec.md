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
)
```
The above constructor returns a `PortSpec` object that defines information about a port for use in [`add_service`](../concepts-reference/subnetworks.md).

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[future-references-reference]: ../concepts-reference/future-references.md
[add-service-reference]: ./plan.md#add_service

---
title: ConnectionConfig
sidebar_label: ConnectionConfig
---

The `ConnectionConfig` object is used to configure a connection between two [subnetworks][subnetworks-reference] (see also, [set_connection][set-connection-reference]).

```python
connection_config = ConnectionConfig(
    # Percentage of packets dropped *each way* between two subnetworks
    # NOTE: for two-way connections like TCP, be conscious that both outbound
    #  and inbound packets will have the drop percentage applied (just like in a real, lossy network link)
    # OPTIONAL (default: 0.0)
    packet_loss_percentage = 50.0,

    # Amount of delay added to packets *each way* between subnetworks
    # Valid values are:
    #  - None
    #  - An instance of UniformPacketDelayDistribution
    #  - An instance of NormalPacketDelayDistribution
    # OPTIONAL (default: None)
    packet_delay_distribution = UniformPacketDelayDistribution(
        # Delay in ms
        ms = 500,
    ),
)
```

:::tip
See [`kurtosis.connection`][connection-prebuilt-enums] for pre-built `ConnectionConfig` objects
:::

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[subnetworks-reference]: ../concepts-reference/subnetworks.md
[set-connection-reference]: ./plan.md#set_connection
[connection-prebuilt-enums]: ./kurtosis.md#connection

---
title: ConnectionConfig
sidebar_label: ConnectionConfig
---

The `ConnectionConfig` is used to configure a connection between two [subnetworks][subnetworks-reference] (see [set_connection][set-connection-reference]).

```python
connection_config = ConnectionConfig(
    # Percentage of packet lost each way between subnetworks 
    # OPTIONAL
    # DEFAULT: 0.0
    packet_loss_percentage = 50.0,

    # Amount of delay added to packets each way between subnetworks
    # OPTIONAL: Valid value are UniformPacketDelayDistribution or NormalPacketDelayDistribution
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

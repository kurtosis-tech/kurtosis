---
title: UniformPacketDelayDistribution
sidebar_label: UniformPacketDelayDistribution
---

The `UniformPacketDelayDistribution` creates a packet delay distribution with constant delay in `ms`. This can be used in conjuction with [`ConnectionConfig`][connection-config] to introduce latency between two [`subnetworks`][subnetworks-reference]. See [`set_connection`][set-connection-reference] instruction to learn more about its usage.

```python

delay  = UniformPacketDelayDistribution(
    # Non-Negative Integer
    # Amount of constant delay added to outgoing packets from the subnetwork
    # MANDATORY
    ms = 1000,
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: ./connection-config.md
[subnetworks-reference]: ../concepts-reference/subnetworks.md
[set-connection-reference]: ./plan.md#set_connection

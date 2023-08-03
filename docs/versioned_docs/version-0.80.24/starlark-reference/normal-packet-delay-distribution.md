---
title: NormalPacketDelayDistribution
sidebar_label: NormalPacketDelayDistribution
---

The `NormalPacketDelayDistribution` creates a packet delay distribution that follows a normal distribution. This can be used in conjunction with [`ConnectionConfig`][connection-config] to introduce latency between two [`subnetworks`][subnetworks-reference]. See [`set_connection`][set-connection-reference] instruction to learn more about its usage.

```python

delay  = NormalPacketDelayDistribution(
    # Non-Negative Integer
    # Amount of mean delay added to outgoing packets from the subnetwork
    # MANDATORY
    mean_ms = 1000,

    # Non-Negative Integer
    # Amount of variance (jitter) added to outgoing packets from the subnetwork
    # MANDATORY
    std_dev_ms = 10,
    
    # Non-Negative Float
    # Percentage of correlation observed among packets. It means that the delay observed in next packet
    # will exhibit a corrlation factor of 10.0% with the previous packet. 
    # OPTIONAL
    # DEFAULT = 0.0
    correlation = 10.0,
)   
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: ./connection-config.md
[subnetworks-reference]: ../concepts-reference/subnetworks.md
[set-connection-reference]: ./plan.md#set_connection

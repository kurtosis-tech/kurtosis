---
title: kurtosis
sidebar_label: kurtosis
---

The `kurtosis` object is an object available in every Kurtosis script that contains prebuilt constants that you might find useful. Each section here represents a property of the `kurtosis` object (e.g. `kurtosis.connection`).

connection
----------

#### `ALLOWED`

`kurtosis.connection.ALLOWED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `0` and `packet_delay` set to `PacketDelay(delay_ms=0)`. It represents a [ConnectionConfig][connection-config] that _allows_ all connection between two [subnetworks][subnetworks-reference] with no delay and packet loss.

#### `BLOCKED`

`kurtosis.connection.BLOCKED` is equivalent to [ConnectionConfig][connection-config] with `packet_loss_percentage` set to `100` and `packet_delay` set to `PacketDelay(delay_ms=0)`. It represents a [ConnectionConfig][connection-config] that _blocks_ all connection between two [subnetworks][subnetworks-reference].

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[connection-config]: ./connection-config.md
[subnetworks-reference]: ../concepts-reference/subnetworks.md

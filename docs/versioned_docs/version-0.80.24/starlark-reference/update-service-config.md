---
title: UpdateServiceConfig
sidebar_label: UpdateServiceConfig
---

The `UpdateServiceConfig` contains the attributes of [ServiceConfig][service-config] that are live-updatable. For now, only the [`subnetwork`][subnetworks-reference] attribute of a service can be updated once the service is started.

```python
update_service_config = UpdateServiceConfig(
    # The subnetwork to which the service will be moved.
    # "default" can be used to move the service to the default subnetwork
    # MANDATORY
    subnetwork = "subnetwork_1",
)
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[service-config]: ./service-config.md
[subnetworks-reference]: ../concepts-reference/subnetworks.md

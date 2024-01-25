---
title: Toleration
sidebar_label: Toleration
---

The `Toleration` constructor creates a `Toleration` object that represents a Kubernetes [Toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) that
can be used with a [ServiceConfig][service-config] object.

```python
toleration = Toleration(
    key = "key",
    operator = "Equal",
    value = "value",
    effect = "NoSchedule",
    toleration_seconds = 64,
)
```

Note all fields are completely optional and follow the rules as laid out in the Kubernetes doc linked above.

Note you need at least one untainted node to use Kurtosis with Kubernetes.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[service-config]: ./service-config.md
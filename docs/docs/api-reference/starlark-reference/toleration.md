---
title: Toleration
sidebar_label: Toleration
---

The `Toleration` constructor creates a `Toleration` object that represents a Kubernetes [Toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) that
can be used with a [ServiceConfig][service-config] object.

```python
toleration = Toleration(
    # key is the taint key that the toleration applies to. Empty means match all taint keys.
    # If the key is empty, operator must be Exists; this combination means to match all values and all keys.
    # OPTIONAL
    key = "key",

    # operator represents a key's relationship to the value.
    # Valid operators are Exists and Equal. Defaults to Equal.
    # Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
    # OPTIONAL
    operator = "Equal",

	# value is the taint value the toleration matches to.
	# If the operator is Exists, the value should be empty, otherwise just a regular string.
	# OPTIONAL
    value = "value",

	# effect indicates the taint effect to match. Empty means match all taint effects.
	# When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
	# OPTIONAL
    effect = "NoSchedule",

	# toleration_seconds represents the period of time the toleration (which must be
	# of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default,
	# it is not set, which means tolerate the taint forever (do not evict). Zero and
	# negative values will be treated as 0 (evict immediately) by the system.
	# OPTIONAL
    toleration_seconds = 64,
)
```

Note all fields are completely optional and follow the rules as laid out in the Kubernetes doc linked above.

Note you need at least one untainted node to use Kurtosis with Kubernetes.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[service-config]: ./service-config.md
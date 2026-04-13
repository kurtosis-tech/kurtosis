---
title: GpuConfig
sidebar_label: GpuConfig
---

The `GpuConfig` constructor creates a `GpuConfig` object that encapsulates GPU-related configuration for a service, including device selection, shared memory size, ulimits, and driver settings. It is used with a [ServiceConfig][service-config] object.

```python
gpu_config = GpuConfig(
    # The number of GPUs to expose to the container.
    # Use -1 to expose all available GPUs, 0 for none, or a positive integer for a specific count.
    # On Docker, sets the DeviceRequests count for the configured driver.
    # On Kubernetes, sets a resource limit/request using the configured k8s resource name
    # (e.g. "nvidia.com/gpu"); only positive counts are supported on Kubernetes.
    # Cannot be used together with device_ids.
    # OPTIONAL (Default: 0)
    count = 2,

    # A list of specific GPU device IDs to pin to the container.
    # Use this when you need to assign particular GPUs by their device index or UUID
    # (e.g. ["0", "1"] or ["GPU-abc123"]).
    # On Docker, sets the DeviceRequests device IDs for the configured driver.
    # NOTE: device_ids is not supported on Kubernetes — use count instead.
    # Cannot be used together with count.
    # OPTIONAL (Default: [])
    device_ids = ["0", "1"],

    # The size of /dev/shm in megabytes.
    # On Docker, sets HostConfig.ShmSize (converted to bytes).
    # On Kubernetes, mounts a memory-backed emptyDir volume at /dev/shm of the given size.
    # OPTIONAL (Default: 0, which uses the runtime default — typically 64 MB)
    shm_size = 128,

    # Resource limits (ulimits) to apply to the container.
    # Keys are ulimit names (e.g. "memlock", "nofile") and values are the limit (soft=hard).
    # Use -1 for unlimited. Common: memlock=-1 for CUDA unified memory.
    # On Docker, maps to HostConfig.Ulimits.
    # NOTE: ulimits are not supported on Kubernetes.
    # OPTIONAL (Default: {})
    ulimits = {
        "memlock": -1,
        "nofile": 65536,
    },

    # The GPU driver to use. Accepts either:
    #   - A string shorthand, e.g. "nvidia" or "amd". The Kubernetes resource name is
    #     derived automatically as "<driver>.com/gpu" (e.g. "nvidia.com/gpu").
    #   - A dict with "docker" and/or "kubernetes" keys for explicit per-backend control,
    #     e.g. {"docker": "amd", "kubernetes": "amd.com/gpu"} or
    #     {"kubernetes": "gpu.intel.com/i915"} (unset keys use their defaults).
    # OPTIONAL (Default: "nvidia")
    driver = "nvidia",
)
```

The above constructor returns a `GpuConfig` object that defines GPU configuration for use in [`ServiceConfig`][service-config] via the `gpu` field.

:::note Kubernetes limitations
- `device_ids` is not supported on Kubernetes — a warning is logged and the field is ignored. Use `count` instead.
- `count = -1` (all GPUs) is not supported on Kubernetes — only positive counts are accepted.
- `ulimits` are not supported on Kubernetes — a warning is logged and the field is ignored.
:::

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[service-config]: ./service-config.md

---
title: Running privileged containers, host PID namespace, and Docker-socket services
sidebar_label: Privileged containers & Docker socket
slug: /running-privileged-containers
sidebar_position: 17
---

This guide covers three related, **Docker-only**, opt-in `ServiceConfig` features:

- `privileged` — start a container with Docker's `--privileged` flag.
- `bind_mounts` — bind-mount allowlisted host paths into the container. The only host path currently allowlisted is `/var/run/docker.sock`, which lets a container talk to the host's Docker daemon (a "docker-in-docker"–style service such as [`ethpandaops/disruptoor`](https://github.com/ethpandaops/disruptoor)).
- `host_pid_namespace` — start a container with Docker's `--pid=host`, which lets tools such as disruptoor use `nsenter` against sibling container network namespaces.

Both features grant the resulting container elevated access to the host. They are denied by default and must be explicitly allowed for the run that interprets the Starlark. Use them only with images you trust.

## When to use this

Reach for these settings only if the workload genuinely needs them. Typical cases:

- A service that must control its own networking primitives that aren't reachable via Linux capabilities alone.
- A controller-style service that orchestrates other containers via the host Docker daemon (e.g. fault-injection or chaos tooling that creates / kills sibling containers).

If you only need a specific Linux capability (`NET_ADMIN`, `SYS_PTRACE`, …), prefer the `capabilities` field — it's narrower and works on both backends.

## Limitations

- **Docker only.** These fields are rejected at interpretation time on the Kubernetes backend with an explicit error.
- **`bind_mounts` host paths are allowlisted.** Today only `/var/run/docker.sock` is permitted. Attempting any other host path (`/etc/passwd`, `/`, `/home/...`, …) fails at interpretation time. Expanding the allowlist is a deliberate code change in `kurtosis_types/service_config/service_config.go`, not configuration.
- **Run opt-in required.** A package or JSON service config can request `privileged=True`, `bind_mounts={...}`, or `host_pid_namespace=True`, but the run must opt in with `--privileged`, `allow-privileged-mode: true`, or the API's `allow_privileged_mode` field.
- **No clear operation yet.** `plan.set_service` and `kurtosis service update` preserve and can enable privileged fields, but cannot currently clear `privileged=True`, remove existing bind mounts, or unset `host_pid_namespace=True`.

## Example

```python
def run(plan):
    plan.add_service(
        name = "disruptoor",
        config = ServiceConfig(
            image = "ethpandaops/disruptoor:latest",

            # Run with Docker's --privileged flag. DANGEROUS — the container can
            # escape isolation. Only use with trusted images.
            privileged = True,

            # Mount the host Docker socket into the container so it can talk to
            # the host Docker daemon. host_path -> container_path. Only
            # /var/run/docker.sock is allowed as a host path.
            bind_mounts = {
                "/var/run/docker.sock": "/var/run/docker.sock",
            },

            # Share the host PID namespace so disruptoor can nsenter target
            # service network namespaces via /proc/<pid>/ns/net.
            host_pid_namespace = True,
        ),
    )
```

If you run this package without opting in, interpretation fails before execution:

```
ServiceConfig requested privileged=true, bind_mounts, or host_pid_namespace=true, but this run did not opt in.
Pass --privileged on the CLI, or set allow-privileged-mode: true in kurtosis-config.yml
```

If you try this on the Kubernetes backend, interpretation fails even if the run opts in:

```
ServiceConfig requested privileged=true, bind_mounts, or host_pid_namespace=true,
but these settings are Docker-only and are not supported on the Kubernetes backend
```

If you try to bind-mount a host path that isn't allowlisted, the package fails at interpretation time before any container is started:

```
'bind_mounts' host path "/etc/passwd" is not permitted; only the following
host paths are allowed: [/var/run/docker.sock]
```

## Allowing a run

For a one-off CLI run, pass `--privileged`:

```bash
kurtosis run --privileged github.com/example/package
```

For `kurtosis service add`, the flag permits a JSON service config that explicitly contains `privileged: true`, `bind_mounts`, or `host_pid_namespace: true`:

```bash
kurtosis service add --privileged my-enclave disruptoor --json-service-config ./service-config.json
```

For `kurtosis service update`, the flag permits updating a service that already has privileged fields so those fields can be preserved:

```bash
kurtosis service update --privileged my-enclave disruptoor --image ethpandaops/disruptoor:latest
```

To allow privileged mode for all CLI runs against one configured cluster, set `allow-privileged-mode: true` under that cluster in `kurtosis-config.yml`:

```yaml
config-version: 8
kurtosis-clusters:
  docker:
    type: docker
    allow-privileged-mode: true
```

For direct API/SDK usage, set the run request's `allow_privileged_mode` field or use the SDK's privileged-mode run config option.

`--privileged` and `allow-privileged-mode` are allow flags. They do not make every service privileged. A service only receives elevated access if its `ServiceConfig` explicitly sets `privileged=True`, `bind_mounts`, or `host_pid_namespace=True`.

When a privileged, bind-mounted, or host-PID service starts, Kurtosis logs a warning recording which service is being granted what, which is useful for auditing in CI logs.

## Security model

- The allowlist is enforced before execution when the Starlark value is converted to a backend `ServiceConfig`.
- The config value is read by the CLI and forwarded to APIC as an `allow_privileged_mode` request field. It is not an engine-side operator ceiling; direct API clients can opt in by setting the API field themselves.
- Privileged, bind-mounted, and host-PID containers can pierce normal container isolation. Treat package authors who use these flags with the same level of trust you'd give a host-level script.

## See also

- [`ServiceConfig` reference](../api-reference/starlark-reference/service-config.md) — full field-by-field documentation including `privileged`, `bind_mounts`, and `host_pid_namespace`.
- [Linux `capabilities(7)`](https://man7.org/linux/man-pages/man7/capabilities.7.html) — for narrower alternatives to `privileged`.

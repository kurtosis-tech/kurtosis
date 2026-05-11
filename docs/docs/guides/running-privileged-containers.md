---
title: Running privileged containers and Docker-socket services
sidebar_label: Privileged containers & Docker socket
slug: /running-privileged-containers
sidebar_position: 17
---

This guide covers two related, **Docker-only**, opt-in `ServiceConfig` features:

- `privileged` — start a container with Docker's `--privileged` flag.
- `bind_mounts` — bind-mount allowlisted host paths into the container. The only host path currently allowlisted is `/var/run/docker.sock`, which lets a container talk to the host's Docker daemon (a "docker-in-docker"–style service such as [`ethpandaops/disruptoor`](https://github.com/ethpandaops/disruptoor)).

Both features grant the resulting container elevated access to the host. They are gated by an allowlist at interpretation time and by an engine-level kill switch at runtime. Use them only with images you trust.

## When to use this

Reach for these settings only if the workload genuinely needs them. Typical cases:

- A service that must control its own networking primitives that aren't reachable via Linux capabilities alone.
- A controller-style service that orchestrates other containers via the host Docker daemon (e.g. fault-injection or chaos tooling that creates / kills sibling containers).

If you only need a specific Linux capability (`NET_ADMIN`, `SYS_PTRACE`, …), prefer the `capabilities` field — it's narrower and works on both backends.

## Limitations

- **Docker only.** Both fields are rejected at service-start time on the Kubernetes backend with an explicit error. There is no Kubernetes equivalent planned in this PR.
- **`bind_mounts` host paths are allowlisted.** Today only `/var/run/docker.sock` is permitted. Attempting any other host path (`/etc/passwd`, `/`, `/home/...`, …) fails at interpretation time. Expanding the allowlist is a deliberate code change in `kurtosis_types/service_config/service_config.go`, not configuration.
- **Engine kill switch.** Operators of shared or CI Kurtosis instances can disable both features fleet-wide; see [Operator controls](#operator-controls) below.

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
        ),
    )
```

If you try this on the Kubernetes backend, the service will fail to start with an error of the form:

```
Service '...' has privileged=true but privileged containers are not supported
on the Kubernetes backend; this feature is Docker-only
```

If you try to bind-mount a host path that isn't allowlisted, the package will fail at interpretation time before any container is started:

```
'bind_mounts' host path "/etc/passwd" is not permitted; only the following
host paths are allowed: [/var/run/docker.sock]
```

## Operator controls

The engine reads `KURTOSIS_ALLOW_PRIVILEGED_CONTAINERS` at service-start time:

- **Unset** (default): `privileged` and `bind_mounts` are allowed.
- **`false`** (case-insensitive, whitespace-tolerant): any service requesting either feature is rejected with an error pointing at this env var. The rest of the package continues to work normally.
- **Any other value** (including `true`): allowed.

To lock down a shared engine, set the env var on the engine process before starting it:

```bash
KURTOSIS_ALLOW_PRIVILEGED_CONTAINERS=false kurtosis engine start
```

The error a package author will see against a locked-down engine looks like:

```
service '<name>' requested privileged=true or bind_mounts but the engine has
KURTOSIS_ALLOW_PRIVILEGED_CONTAINERS=false; restart the engine without that
environment variable (or set it to true) to enable privileged containers
```

When a privileged or bind-mounted service does start on a permissive engine, the engine logs a warning recording which service is being granted what — useful for auditing in CI logs.

## Security model

- The allowlist is enforced **twice**: once as the `bind_mounts` attribute validator at interpretation time, and again inside `ToKurtosisType` when the Starlark value is converted to a backend `ServiceConfig`. This is intentional belt-and-suspenders; either layer alone would be sufficient.
- The kill switch defaults to **allowed** (fail-open) so existing deployments aren't broken by an upgrade. Operators who care about lock-down must opt in by setting the env var to `false`.
- Privileged and bind-mounted containers are functionally equivalent to handing the workload root access on the host. Treat package authors who use these flags with the same level of trust you'd give a host-level script.

## See also

- [`ServiceConfig` reference](../api-reference/starlark-reference/service-config.md) — full field-by-field documentation including `privileged` and `bind_mounts`.
- [Linux `capabilities(7)`](https://man7.org/linux/man-pages/man7/capabilities.7.html) — for narrower alternatives to `privileged`.

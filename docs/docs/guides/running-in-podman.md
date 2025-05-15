---
title: Running Kurtosis in Podman
sidebar_label: Running in Podman
slug: /podman
sidebar_position: 7
---

This guide assumes that you have [Kurtosis installed](../get-started/installing-the-cli.md).

If you would like more information on Podman and how to set up and manage a Podman environment, check out these official [docs](https://docs.podman.io/en/latest/).

I. Set Up Podman
-----------------

1. Install Podman on your system following the [official installation guide](https://docs.podman.io/en/latest/installation.html).

2. Start the Podman socket service:
```bash
podman machine init
podman machine start
```

3. Configure Docker CLI to use Podman by setting the `DOCKER_HOST` environment variable:
```bash
export DOCKER_HOST="unix://$HOME/.local/share/containers/podman/machine/podman.sock"
```

:::tip Docker Compatibility
Podman provides a Docker-compatible CLI and API. This means you can use Docker commands with Podman by either:
- Setting the `DOCKER_HOST` environment variable as shown above
- Using the `podman` command directly (it accepts the same commands as `docker`)
- Using the `podman-docker` package which creates a symlink from `docker` to `podman`

For more information on Docker compatibility, see the [Podman documentation](https://docs.podman.io/en/latest/markdown/podman-run.1.html#docker-compatibility).
:::

II. Add Podman Cluster to `kurtosis-config.yml`
--------------------------------

1. Open the file located at `"$(kurtosis config path)"`. This should look like `/Users/<YOUR_USER>/Library/Application Support/kurtosis/kurtosis-config.yml` on MacOS.

2. Add a new cluster configuration for Podman. The configuration is identical to Docker since Podman is Docker-compatible:

```yaml
config-version: 6
should-send-metrics: true
kurtosis-clusters:
  docker:
    type: "docker"
    ...
  podman:
    type: "podman"
    ...
```

III. Configure Kurtosis
--------------------------------

Run `kurtosis cluster set podman`. This will start the engine using Podman as the container runtime.

Done! Now you can run any Kurtosis command or package just like if you were doing it locally with Docker.

:::tip Switching Back
To switch back to using Kurtosis with Docker, simply use: `kurtosis cluster set docker`
:::
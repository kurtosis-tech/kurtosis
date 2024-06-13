---
title: Running Kurtosis in your own Cloud
sidebar_label: Running in your own Cloud
slug: /self-hosting
sidebar_position: 14
---

This guide will help you set up Kurtosis in your own cloud.

I. Prerequisites
-----------------

1. Public facing gateway supporting your domain with certificate. The certificate should support your domain name and a wildcard subdomain prefix `*.domain` since the service port URLs format is `port-service-enclave.domain`. The gateway should terminate TLS.
2. Host running Ubuntu to install and configure Kurtosis. The host should be on a private subnet receiving traffic from the Gateway on port 80. Healthchecks should use the `/status` URL.

![overview](/img/guides/self-hosting-overview.png)

II. Kurtosis Installation
-----------------

We provide an install script setting up Docker, Nginx and Kurtosis. The script takes as arguments your domain name, a username and password for HTTP basic authentication.

```bash
curl -s https://raw.githubusercontent.com/kurtosis-tech/kurtosis-cloud-config/main/self-hosting-setup.sh | bash -s <domain name> <username> <password>
```

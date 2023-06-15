---
title: service add
sidebar_label: service add
slug: /service-add
---

To add a service to an enclave, run:

```bash
kurtosis service add $THE_ENCLAVE_IDENTIFIER $THE_SERVICE_IDENTIFIER $CONTAINER_IMAGE
```

where `$THE_ENCLAVE_IDENTIFIER` and the `$THE_SERVICE_IDENTIFIER` are [resource identifiers](../concepts-reference/resource-identifier.md) for the enclave and service, respectively. 

Much like `docker run`, this command has multiple options available to customize the service that's started:

1. The `--entrypoint` flag can be passed in to override the binary the service runs
1. The `--env` flag can be used to specify a set of environment variables that should be set when running the service
1. The `--ports` flag can be used to set the ports that the service will listen on

To override the service's CMD, add a `--` after the image name and then pass in your CMD args like so:

```bash
kurtosis service add --entrypoint sh my-enclave test-service alpine -- -c "echo 'Hello world'"
```
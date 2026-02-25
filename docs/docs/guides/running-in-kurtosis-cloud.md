---
title: Running Kurtosis in Kurtosis Cloud
sidebar_label: Running in Kurtosis Cloud
slug: /cloud
sidebar_position: 7
---

Kurtosis Cloud is a fully managed cloud service that provides self-service workflows, allowing for an easy and stress-free deployment of test and dev environments directly onto remote infrastructure. These environments can persist for as long as you require. By logging into our [cloud portal](https://cloud.kurtosis.com), a cloud instance will be provisioned to run your test and dev enclaves.

You can interact with your enclaves using the UI (or the [CLI](../get-started/installing-the-cli.md#ii-install-the-cli) for more advanced use cases).

![enclave-manager-ui](/img/guides/enclave-manager-ui.png)

A Kurtosis cloud instance is an AWS EC2 instance running the Kurtosis engine, the Kurtosis API controller and your enclave services inside Docker.  The service ports are forwarded to your local machine.

![cloud-arch](/img/guides/cloud-arch.png)

### Advantages of running Kurtosis enclaves in Kurtosis Cloud

In addition to offloading the compute to a cloud infrastructure, Kurtosis Cloud comes with other advantages.
When provisioning a cloud instance, Kurtosis will create a specific AWS user account and a storage space in S3.

Services running inside a Kurtosis enclave in Kurtosis Cloud can freely read and write objects from/to this S3 storage.

:::warning S3 storage is publicly accessible
While only the owner of the S3 storage is permitted to write to the S3 storage, the data is publicly accessible to 
anyone that knows the object key. For this reason, we don't recommend storing sensitive data in the S3 storage.
:::

The AWS user key as well as the information on the user S3 space is provided to all Starlark packages running in the 
cloud via the global `kurtosis` module. The following variables are available inside Starlark and can be passed as 
environment variables or simple arguments to the services started inside the enclave:
- `kurtosis.aws_access_key_id` and `kurtosis.aws_secret_access_key`: The AWS user key pair required to authenticate to 
AWS
- `kurtosis.aws_bucket_region`, `kurtosis.aws_bucket_name` and `kurtosis.aws_bucket_user_folder`: The bucket region and 
name, as well as the specific user folder the AWS user is authorized to access.

:::tip AWS user permissions
The AWS user created has a very restricted set of permissions by default. It can only read and write to its user folder
inside the S3 bucket. But it is possible to request more access, reach out to us via [Discord](https://discord.gg/6Jjp9c89z9).
:::

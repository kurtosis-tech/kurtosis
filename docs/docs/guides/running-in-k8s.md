---
title: Running Kurtosis in Kubernetes
sidebar_label: Running in Kubernetes
slug: /k8s
sidebar_position: 6
---

This guide assumes that you have [Kurtosis installed](../get-started/installing-the-cli.md).

If you would like more information on Kubernetes and how to set up, run and manage a cluster check out these official [docs](https://kubernetes.io/docs/home/). 

Please note that in order to ensure Kurtosis works the same way over Kubernetes as it does over Docker locally, service names must be a valid [RFC-1035 Label Name](https://kubernetes.io/docs/advanced-concepts/overview/working-with-objects/names/#rfc-1035-label-names). This means service names must contain: at most 63 characters, only lowercase alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character. 

I. Create a Kubernetes Cluster
-----------------

There are many different ways to get a Kubernetes cluster (roughly ordered easiest to hardest):

- Use [Kubernetes provided with Docker Desktop](https://docs.docker.com/desktop/kubernetes/)
- Install [Minikube](https://github.com/kubernetes/minikube)
- Use [k3s](https://k3s.io/)
- Deploy it on an onprem cluster and manage the machine provisioning yourself
- Deploy it on the cloud, managing the Kubernetes nodes on cloud instances yourself (e.g. EC2, AVM, GCE, etc)
- Deploy it on a managed Kuberenetes cluster, managing scaling and configurations yourself (e.g. EKS, AKS, GKE)

:::tip Kurtosis Kloud Early Access
If you're looking to run a stress-free "Kurtosis on Kubernetes in the cloud", look no further! We're excited to launch an early access offering for [Kurtosis Kloud](https://mp2k8nqxxgj.typeform.com/to/U1HcXT1H). Once you [sign up](https://mp2k8nqxxgj.typeform.com/to/U1HcXT1H), we'll reach out to you with the next steps.
:::


II. Add you Kubernetes Cluster credentials to your `kubeconfig`
-------------------------

This step will depend highly on how your cluster was created. But generally you will need to either:

- Manually edit the `kubeconfig` file to contain cluster and authentication data. For more information, see [Kubernetes docs](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/).
- Use your cloud provider's CLI to automatically edit the `kubeconfig` file so that it contains your cluster and authentication data. For example, you if you are using Amazon's managed Kubernetes service (called EKS), [this tutorial](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html) is comprehensive.


III. Add your cluster information to `kurtosis-config.yml`
--------------------------------

1. Open the file located at `"$(kurtosis config path)"`. This should look like `/Users/<YOUR_USER>/Library/Application Support/kurtosis/kurtosis-config.yml` on MacOS.
2. Paste the following contents, changing `NAME-OF-YOUR-CLUSTER` to the cluster you created and save:
```yaml
config-version: 2
should-send-metrics: true
kurtosis-clusters:
  docker:
    type: "docker"
  minikube:
    type: "kubernetes"
    config:
      kubernetes-cluster-name: "minikube"
      storage-class: "standard"
      enclave-size-in-megabytes: 10
  cloud:
    type: "kubernetes"
    config:
      kubernetes-cluster-name: "NAME-OF-YOUR-CLUSTER"
      storage-class: "standard"
      enclave-size-in-megabytes: 10
```

IV. Configure Kurtosis
--------------------------------

1. Run `kurtosis cluster set cloud`.  This will start the engine remotely. See the CLI reference for more information about `kurtosis cluster` commands [here](../cli-reference/cluster-set.md).
1. *In another terminal*, run [`kurtosis gateway`](../cli-reference/gateway.md). This will act as a middle man between your computer's ports and your services deployed on Kubernetes ports and has to stay running as a separate process.

Done! Now you can run any Kurtosis command or package just like if you were doing it locally.

:::tip Kurtosis Kloud Early Access
To switch back to using Kurtosis locally, simply use: `kurtosis cluster set docker`
:::


V. \[Optional] Activate the enclave pool to accelerate the enclave creation time
--------------------------------

This step is optional, but we recommend taking it as it improves the user experience during the enclave creation, specifically regarding speed.

Creating a new enclave from scratch demands several time-consuming engine tasks and the creation of resources.

The enclave pool feature was introduced to reduce the time it takes for a user to run a Kurtosis package in the cloud by spinning up the enclaves before they are needed.

The enclave pool is a functionality of the Kurtosis engine that automatically creates `idle` enclaves, when the engine is started, that are then used whenever users need to create a new enclave (e.g: when running the `kurtosis enclave add` command).

This mechanism reduces enclave creation time by using a running `idle` enclave when a new enclave is requested from the engine.

To enable this feature you have to run the following:

1. Run `kurtosis engine restart --enclave-pool-size {pool-size-number}`. If you already follow the previous step and replace the {pool-size-number} with an integer

OR

1. Run `kurtosis engine start --enclave-pool-size {pool-size-number}`. If the engine has not been started yet.

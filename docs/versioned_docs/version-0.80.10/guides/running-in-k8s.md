---
title: Running Kurtosis in Kubernetes
sidebar_label: Running in Kubernetes
slug: /k8s
---

This guide assumes that you have [Kurtosis installed](./installing-the-cli.md).

If you would like more information on Kubernetes and how to set up, run and manage a cluster check out these offical [docs](https://kubernetes.io/docs/home/)

I. Create a Kubernetes Cluster
-----------------

There are many diferent ways to get a Kubernetes cluster (roughly ordered easiest to hardest):

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
1. Paste the following contents, changing `NAME-OF-YOUR-CLUSTER` to the cluster you created and save:
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

1. Run `kurtosis cluster set cloud`.  This will start the engine remotely. 
1. *In another terminal*, run `kurtosis gateway`. This will act as a middle man between your computer's ports and your services deployed on Kubernetes ports and has to stay running as a separate proccess.

Done! Now you can run any Kurtosis command or package just like if you were doing it locally.

:::tip Kurtosis Kloud Early Access
To switch back to using Kurtosis locally, simply use: `kurtosis cluster set docker`
:::

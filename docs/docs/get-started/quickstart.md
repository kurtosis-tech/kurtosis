---
title: Quickstart
id: quickstart
sidebar_label: Quickstart
slug: /quickstart
toc_max_heading_level: 2
sidebar_position: 3
---

Introduction
------------

Welcome to the [Kurtosis][homepage] quickstart! This guide takes ~5 minutes and will walk you through running a containerized application in Kurtosis using the Kurtosis CLI. To see how to write your own package, check out our [quickstart on writing a package.][how-to-set-up-postgres-guide]

:::tip What You'll Do
1. Install Kurtosis, if you haven't already.
2. Run a basic application directly from a Kurtosis package hosted on Github.
3. Inspect your environment with the Kurtosis CLI.
4. Modify your environment by passing in arguments to the package in JSON format.
5. Run the application over Kubernetes, instead of Docker.
:::

<details><summary>Forget installing! Let me do it on Gitpod</summary>

If you don't want to install Docker and Kurtosis, you can run through the quickstart on Gitpod:
 
**Open the Playground: [Start](https://gitpod.io/?autoStart=true&editor=code#https://github.com/kurtosis-tech/ethereum-package)**

Click on the "New Workspace" button! You don't have to worry about the Context URL, Editor or Class. It's all pre-configured for you.
 
</details>

1. Install dependencies
--------------------
Before you get started, make sure you have:
* [Installed Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal to see all your Docker images.
* [Installed Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

2. Run a basic package from Github
---------------------------------------

Run the following in your CLI:

```console
kurtosis run github.com/kurtosis-tech/basic-service-package
```

You should get output that looks like:

![basic-service-default-output.png](/img/home/basic-service-default-output.png)

Spin up your system!
--------------------
Great! You're now ready to bring up your own network. Simply run:
```console
kurtosis run github.com/kurtosis-tech/ethereum-package --args-file ~/network_params.yaml --enclave eth-network
```

Kurtosis will then begin to spin up your private Ethereum testnet by interpreting the instructions in the Kurtosis package, validating the plan to ensure there are no conflicts or obvious errors, and then finally executes the plan (read more about multi-phase runs [here][multi-phase-runs-reference]). Kurtosis first spins up an isolated, ephemeral environment on your machine called an [enclave][enclaves-reference] where all the services and [files artifacts][files-artifacts-reference] for your system will reside in. Then, those services will be bootstrapped and required files generated to start up the system.

You will see a stream of text get printed in your terminal as Kurtosis begins to generate genesis files, configure the Ethereum nodes, launch a Grafana and Prometheus instance, and bootstrap the network together. In ~2 minutes, you should see the following output at the end:

```console
INFO[2023-08-28T13:05:31-04:00] ====================================================
INFO[2023-08-28T13:05:31-04:00] ||          Created enclave: eth-network          ||
INFO[2023-08-28T13:05:31-04:00] ====================================================
Name:            eth-network
UUID:            e1a41707ee8e
Status:          RUNNING
Creation Time:   Mon, 28 Aug 2023 13:04:53 EDT

========================================= Files Artifacts =========================================
UUID           Name
a662c7c74685   1-lighthouse-geth-0-63
6421d80946ce   2-lighthouse-geth-64-127
a1ad3962f148   cl-genesis-data
730d585d5ec5   el-genesis-data
c1e452ad7e53   genesis-generation-config-cl
284cde692102   genesis-generation-config-el
b03a5b7b9340   geth-prefunded-keys
013f5d8708fa   prysm-password

========================================== User Services ==========================================
UUID           Name                                       Ports                                         Status
202516f0ff8f   cl-1-lighthouse-geth                       http: 4000/tcp -> http://127.0.0.1:65191      RUNNING
                                                          metrics: 5054/tcp -> http://127.0.0.1:65192
                                                          tcp-discovery: 9000/tcp -> 127.0.0.1:65193
                                                          udp-discovery: 9000/udp -> 127.0.0.1:64174
66bdfbd6c066   cl-1-lighthouse-geth-validator             http: 5042/tcp -> 127.0.0.1:65236             RUNNING
                                                          metrics: 5064/tcp -> http://127.0.0.1:65237
b636913d4d03   cl-2-lighthouse-geth                       http: 4000/tcp -> http://127.0.0.1:65311      RUNNING
                                                          metrics: 5054/tcp -> http://127.0.0.1:65312
                                                          tcp-discovery: 9000/tcp -> 127.0.0.1:65310
                                                          udp-discovery: 9000/udp -> 127.0.0.1:51807
e296eefa1710   cl-2-lighthouse-geth-validator             http: 5042/tcp -> 127.0.0.1:65427             RUNNING
                                                          metrics: 5064/tcp -> http://127.0.0.1:65428
4df1beb0203d   el-1-geth-lighthouse                       engine-rpc: 8551/tcp -> 127.0.0.1:65081       RUNNING
                                                          rpc: 8545/tcp -> 127.0.0.1:65079
                                                          tcp-discovery: 30303/tcp -> 127.0.0.1:65078
                                                          udp-discovery: 30303/udp -> 127.0.0.1:55146
                                                          ws: 8546/tcp -> 127.0.0.1:65080
581a0fe5de77   el-2-geth-lighthouse                       engine-rpc: 8551/tcp -> 127.0.0.1:65130       RUNNING
                                                          rpc: 8545/tcp -> 127.0.0.1:65132
                                                          tcp-discovery: 30303/tcp -> 127.0.0.1:65131
                                                          udp-discovery: 30303/udp -> 127.0.0.1:49475
                                                          ws: 8546/tcp -> 127.0.0.1:65129
4980884d9bb0   prelaunch-data-generator-cl-genesis-data   <none>                                        RUNNING
3174baf6a6ff   prelaunch-data-generator-el-genesis-data   <none>                                        RUNNING
```

Thats it! You now have a full, private Ethereum blockchain on your local machine.

The first section that gets printed contains some basic metadata about the enclave that was spun up. This includes the name of the enclave `eth-network`, its [Resource Idenfitier](https://docs.kurtosis.com/advanced-concepts/resource-identifier/), your enclave's status, and the time it was created.

Next, you'll see a section dedicated to [Files Artifacts](https://docs.kurtosis.com/advanced-concepts/files-artifacts/), which are Kurtosis' first-class representation of data inside your enclave, stored as compressed TGZ files. You'll notice there are configuration files for the nodes, grafana, and prometheus as well as private keys for pre-funded accounts and genesis-related data. These files artifacts were generated and used by Kurtosis to start the network and abstracts away the complexities and overhead that come with generating validator keys and getting genesis and node config files produced and mounted to the right containers yourself.

Lastly, there is a section called `User Services` which display the number of services (running in Docker containers) that make up your network. You will notice that there are 2 Ethereum nodes comprised of 3 services each (an EL client, a CL beacon client, and a CL validator client) and 2 genesis data generators for each the CL and EL. Each of these services are running in Docker containers inside your local enclave & Kurtosis has automatically mapped each container port to your machine's ephemeral ports for seamless interaction with the services running in your enclave.

Why Kurtosis packages - from a consumer's perspective
-----------------------------------------------------
Kurtosis was built to make building distributed systems as easy as building a single server app. Kurtosis aims to achieve this by bridging the environment definition author-consumer divide. Tactically, this means making it dead simple for a consumer (like yourself) to pick up an environment definition, spin it up, and deploy it the way you want, where you want - all without needing to know specialized knowledge about how the system works or how to use Kubernetes or Docker. 

Specifically, this guide showed you:
- ***The power of parameterizability***: as a consumer of the environment definition, having both the knowledge and means to configure the system to spin up the way you need it is incredibly valuable - a big reason why Kurtosis packages are meant to be parameterized. In this guide, you created the `network_params.yaml` file which contained your preferences for how the network should look and passed them in to Kurtosis with relative ease. The author of the package need only define the arguments and flags available for a consumer, and Kurtosis handles the rest once those are passed in at runtime.
- ***Portable and easy to wield***: a major contributor to the author-consumer divide comes from the knowledge gap between the author and consumer regarding the infrastruture and tools needed to instantiate a system. Understanding how Kubernetes works, what Bash script to use at which step, and working with Docker primitivies are all pain points we believe Kurtosis alleviates. In this guide, you installed Kurtosis and ran a single command to get your system up and running. This same command will work anywhere, over Docker or on Kubernetes, locally or on remote infrastructure. We believe this portability and ease of use are requirements for bridging the author-consumer divide.

There are many other reasons why we believe Kurtosis is the right tool for bridging the author-consumer divide. Check out the [next guide][how-to-set-up-postgres-guide] to experience the workflow for a package author and how Kurtosis improves the developer experience for an environment definition author.

Conclusion
----------
And that's it - you've successfully used Kurtosis to instantiate a full, private Ethereum testnet - one of the most complex distributed systems in todays time.

Let's review. In this tutorial you have:

1. Installed Kurtosis and Docker.
2. Configure how your system should look like, using parameters that are passed in at runtime.
3. Run a single command to spin up your network.
4. Reviewed how package consumers benefit from using environment definitions written for Kurtosis.

:::tip
In this short guide, you went through the workflow that a Kurtosis package consumer would experience. It is strongly encouraged that you check out the [next guide][how-to-set-up-postgres-guide] where you will set up a Postgres database and an API server to as a package author.
:::
   
This was still just an introduction to Kurtosis. To dig deeper, visit other sections of our docs where you can read about [what Kurtosis is][homepage], understand the [architecture][architecture-explanation], and hear our [inspiration for starting Kurtosis][why-kurtosis-explanation]. 

To learn more about how Kurtosis is used, we encourage you to check out our [`awesome-kurtosis` repository][awesome-kurtosis-repo], where you will find real-world examples of Kurtosis in action, including:
- How to run a simple [Go test][go-test-example] or [Typescript test][ts-test-example] against the app we just built
- The [Ethereum package][ethereum-package], used by the Ethereum Foundation, which can be used to set up local testnets 
- A parameterized package for standing up an [n-node Cassandra cluster with Grafana and Prometheus][cassandra-package-example] out-of-the-box
- The [NEAR package][near-package] for local dApp development in the NEAR ecosystem

Finally, we'd love to hear from you. Please don't hesitate to share with us what went well, and what didn't, using `kurtosis feedback` to file an issue in our [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) or to [chat with our cofounder, Kevin](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).

Lastly, feel free to [star us on Github](https://github.com/kurtosis-tech/kurtosis), post your questions on our [Github Discussions Forum][github-discussions], [join the community in our Discord](https://discord.gg/6Jjp9c89z9), and [follow us on Twitter](https://twitter.com/KurtosisTech)!

Thank you for trying our quickstart. We hope you enjoyed it. 

<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ../get-started/installing-the-cli.md
[installing-docker-guide]: ../get-started/installing-the-cli.md#i-install--start-docker
[upgrading-kurtosis-guide]: ../guides/upgrading-the-cli.md
[how-to-set-up-postgres-guide]: write-your-first-package.md

<!--------------------------- Advanced Concepts ------------------------------------>
[architecture-explanation]: ../advanced-concepts/architecture.md
[enclaves-reference]: ../advanced-concepts/enclaves.md
[services-explanation]: ../advanced-concepts/architecture.md#services
[reusable-environment-definitions-explanation]: ../advanced-concepts/reusable-environment-definitions.md
[why-kurtosis-explanation]: ../advanced-concepts/why-kurtosis.md
[how-do-imports-work-explanation]: ../advanced-concepts/how-do-kurtosis-imports-work.md
[why-multi-phase-runs-explanation]: ../advanced-concepts/why-multi-phase-runs.md

<!--------------------------- Reference ------------------------------------>
<!-- CLI Commands Reference -->
[cli-reference]: /cli/
[kurtosis-run-reference]: ../cli-reference/run.md
[kurtosis-clean-reference]: ../cli-reference/clean.md
[kurtosis-enclave-inspect-reference]: ../cli-reference/enclave-inspect.md
[kurtosis-files-upload-reference]: ../cli-reference/files-upload.md
[kurtosis-feedback-reference]: ../cli-reference/feedback.md
[kurtosis-twitter]: ../cli-reference/twitter.md
[starlark-reference]: ../advanced-concepts/starlark.md

<!-- SL Instructions Reference-->
[request-reference]: ../api-reference/starlark-reference/plan.md#request
[exec-reference]: ../api-reference/starlark-reference/plan.md#exec

<!-- Reference -->
[multi-phase-runs-reference]: ../advanced-concepts/multi-phase-runs.md
[kurtosis-yml-reference]: ../advanced-concepts/kurtosis-yml.md
[packages-reference]: ../advanced-concepts/packages.md
[runnable-packages-reference]: ../advanced-concepts/packages.md#runnable-packages
[locators-reference]: ../advanced-concepts/locators.md
[plan-reference]: ../advanced-concepts/plan.md
[future-references-reference]: ../advanced-concepts/future-references.md
[files-artifacts-reference]: ../advanced-concepts/files-artifacts.md

<!--------------------------- Other ------------------------------------>
<!-- Examples repo -->
[awesome-kurtosis-repo]: https://github.com/kurtosis-tech/awesome-kurtosis
[data-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/dvd-rental-data.tar
[cassandra-package-example]: https://github.com/kurtosis-tech/cassandra-package
[go-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/go-test
[ts-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/ts-test
[ethereum-package]: https://github.com/kurtosis-tech/ethereum-package/

<!-- Misc -->
[homepage]: get-started.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[bazel-github]: https://github.com/bazelbuild/bazel/
[starlark-github-repo]: https://github.com/bazelbuild/starlark
[postgrest]: https://postgrest.org/en/stable/
[waku-package]: https://github.com/logos-co/wakurtosis
[near-package]: https://github.com/kurtosis-tech/near-package
[iterm]: https://iterm2.com/
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
[github-discussions]: https://github.com/kurtosis-tech/kurtosis/discussions/new?category=q-a

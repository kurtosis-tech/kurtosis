---
title: Quickstart
sidebar_label: Quickstart
slug: /quickstart
toc_max_heading_level: 2
---

Introduction
------------

Welcome to the [Kurtosis][homepage] quickstart! This guide takes ~5 minutes and will walk you through how to use a Kurtosis package to spin up a distributed system over Docker. Specifically, you will use the [eth2-package][eth2-package] to bootstrap and start-up a private Ethereum testnet. 

Kurtosis is a build system for test environments that serve two types of users: the author of the environment definition and the consumer or user of the environment definition. This quickstart is intended to put you in the shoes of the consumer - someone who needs a quick and easy way to get a test environment that represents their system holistically. A separate guide is [available here](TODO) to introduce you to the world of environment definition authoring and how one might define and build an environment with Kurtosis for themselves or for their team to use when developing and testing against a distributed system.

For a quick read on what Kurtosis is and what problems Kurtosis aims to solve, our [introduction page][homepage] will be a great starting point, alongside our [motivations behind starting Kurtosis][why-we-built-kurtosis-explanation].

This guide is in a "code along" format, meaning we assume the user will be following the code examples and running Kurtosis CLI commands on your local machine. Everything you run in this guide is free, public, and does not contain any sensitive data. 

:::tip What You'll Do
1. Install Kurtosis and Docker, if you haven't already.
1. Configure how you want the system to be spun up, using parameters that are passed in at runtime.
1. Run a single command to spin up your network.
:::

<details><summary>TL;DR Version</summary>

This quickstart is in a "code along" format. You can also dive straight into running the end results and exploring the code too.
 
**Open the Playground: [Start](https://gitpod.io/?autoStart=true&editor=code#https://github.com/kurtosis-tech/eth2-package)**

Click on the "New Workspace" button! You don't have to worry about the Context URL, Editor or Class. It's all pre-configured for you.
 
</details>

If you ever get stuck, every Kurtosis command accepts a `-h` flag to print helptext. If that doesn't help, you can get in touch with us in our [Discord server](https://discord.gg/jJFG7XBqcY) or on [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose)!

Setup
-----

#### Requirements
Before you proceed, please make sure you have:
- [Installed and started the Docker engine][installing-docker-guide]
- [Installed the Kurtosis CLI][installing-kurtosis-guide] (or [upgraded to latest][upgrading-kurtosis-guide] if you already have it)

#### Install dependencies
* [Install Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal to see all your Docker images.
* [Install Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

#### Configure your network
Next, create a file titled: `eth2-package-params.json` in your working directory and populate it with:
```json
{
	"participants": [{
		"el_client_type": "geth",
		"el_client_image": "ethereum/client-go:latest",
		"el_client_log_level": "",
		"el_extra_params": [],
		"cl_client_type": "lighthouse",
		"cl_client_image": "sigp/lighthouse:latest",
		"cl_client_log_level": "",
		"beacon_extra_params": [],
		"validator_extra_params": [],
		"builder_network_params": null
	}],
	"network_params": {
		"network_id": "3151908",
		"deposit_contract_address": "0x4242424242424242424242424242424242424242",
		"seconds_per_slot": 12,
		"slots_per_epoch": 32,
		"num_validator_keys_per_node": 64,
		"preregistered_validator_keys_mnemonic": "giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete",
		"deneb_for_epoch": 500
	},
	"verifications_epoch_limit": 5,
	"global_client_log_level": "info",
	"mev_type": "full"
}
```
You will use the above file by passing it in at runtime, effectively enabling you to define the way your network should look using parameters.

#### Launch the network with `full MEV`
Great! You're now ready to bring up your own network. Simply run:
```bash
TODO: FIX kurtosis run --enclave eth-network github.com/kurtosis-tech/eth2-package "$(cat ~/eth2-package-params.json)"
```
Kurtosis will then begin to spin up your private Ethereum testnet with `full MEV`. You will see a stream of text get printed in your terminal as Kurtosis begins to generate genesis files, configure the Ethereum nodes, launch a Grafana and Prometheus instance, and bootstrap the network together with the full suite of MEV products from Flashbots. In ~2 minutes, you should see the following output at the end:

Conclusion
----------
And that's it - you've written your very first distributed application in Kurtosis!

Let's review. In this tutorial you have:

- Started a Postgres database in an ephemeral, isolated test environment
- Seeded your database by importing an external Starlark package from the internet
- Set up an API server for your database and gracefully handled dynamically generated dependency data
- Inserted & queried data via the API
- Parameterized data insertion for future use

This was still just an introduction to Kurtosis. To dig deeper, visit other sections of our docs where you can read about [what Kurtosis is][homepage], understand the [architecture][architecture-explanation], and hear our [inspiration for starting Kurtosis][why-we-built-kurtosis-explanation]. 

To learn more about how Kurtosis is used, we encourage you to check out our [`awesome-kurtosis` repository][awesome-kurtosis-repo], where you will find real-world examples of Kurtosis in action, including:
- How to run a simple [Go test][go-test-example] or [Typescript test][ts-test-example] against the app we just built
- The [Ethereum package][ethereum-package], used by the Ethereum Foundation, which can be used to set up local testnets 
- A parameterized package for standing up an [n-node Cassandra cluster with Grafana and Prometheus][cassandra-package-example] out-of-the-box
- The [NEAR package][near-package] for local dApp development in the NEAR ecosystem

Finally, we'd love to hear from you. Please don't hesitate to share with us what went well, and what didn't, using `kurtosis feedback` to file an issue in our [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) or to [chat with our cofounder, Kevin](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding).

Lastly, feel free to [star us on Github](https://github.com/kurtosis-tech/kurtosis), [join the community in our Discord](https://discord.com/channels/783719264308953108/783719264308953111), and [follow us on Twitter](https://twitter.com/KurtosisTech)!

Thank you for trying our quickstart. We hope you enjoyed it. 

<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ./guides/installing-the-cli.md#ii-install-the-cli
[installing-docker-guide]: ./guides/installing-the-cli.md#i-install--start-docker
[upgrading-kurtosis-guide]: ./guides/upgrading-the-cli.md

<!--------------------------- Explanations ------------------------------------>
[architecture-explanation]: ./explanations/architecture.md
[enclaves-reference]: ./concepts-reference/enclaves.md
[services-explanation]: ./explanations/architecture.md#services
[reusable-environment-definitions-explanation]: ./explanations/reusable-environment-definitions.md
[why-we-built-kurtosis-explanation]: ./explanations/why-we-built-kurtosis.md
[how-do-imports-work-explanation]: ./explanations/how-do-kurtosis-imports-work.md
[why-multi-phase-runs-explanation]: ./explanations/why-multi-phase-runs.md

<!--------------------------- Reference ------------------------------------>
<!-- CLI Commands Reference -->
[cli-reference]: /cli
[kurtosis-run-reference]: ./cli-reference/run.md
[kurtosis-clean-reference]: ./cli-reference/clean.md
[kurtosis-enclave-inspect-reference]: ./cli-reference/enclave-inspect.md
[kurtosis-files-upload-reference]: ./cli-reference/files-upload.md
[kurtosis-feedback-reference]: ./cli-reference/feedback.md
[kurtosis-twitter]: ./cli-reference/twitter.md
[starlark-reference]: ./concepts-reference/starlark.md

<!-- SL Instructions Reference-->
[request-reference]: ./starlark-reference/plan.md#request
[exec-reference]: ./starlark-reference/plan.md#exec

<!-- Reference -->
[multi-phase-runs-reference]: ./concepts-reference/multi-phase-runs.md
[kurtosis-yml-reference]: ./concepts-reference/kurtosis-yml.md
[packages-reference]: ./concepts-reference/packages.md
[runnable-packages-reference]: ./concepts-reference/packages.md#runnable-packages
[locators-reference]: ./concepts-reference/locators.md
[plan-reference]: ./concepts-reference/plan.md
[future-references-reference]: ./concepts-reference/future-references.md
[files-artifacts-reference]: ./concepts-reference/files-artifacts.md

<!--------------------------- Other ------------------------------------>
<!-- Examples repo -->
[awesome-kurtosis-repo]: https://github.com/kurtosis-tech/awesome-kurtosis
[data-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/dvd-rental-data.tar
[cassandra-package-example]: https://github.com/kurtosis-tech/cassandra-package
[go-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/go-test
[ts-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/ts-test
[eth2-package]: https://github.com/kurtosis-tech/eth2-package/

<!-- Misc -->
[homepage]: home.md
[kurtosis-managed-packages]: https://github.com/kurtosis-tech?q=in%3Aname+package&type=all&language=&sort=
[wild-kurtosis-packages]: https://github.com/search?q=filename%3Akurtosis.yml&type=code
[bazel-github]: https://github.com/bazelbuild/bazel/
[starlark-github-repo]: https://github.com/bazelbuild/starlark
[postgrest]: https://postgrest.org/en/stable/
[ethereum-package]: https://github.com/kurtosis-tech/eth2-package
[waku-package]: https://github.com/logos-co/wakurtosis
[near-package]: https://github.com/kurtosis-tech/near-package
[iterm]: https://iterm2.com/
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension

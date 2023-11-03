---
title: How to build your own private, Ethereum testnet with eth-kurtosis
sidebar_label: Build your own testnet from scratch
slug: /how-to-compose-your-own-testnet
toc_max_heading_level: 2
sidebar_position: 11
---

:::tip
If you'd prefer to dive into the code example, visit the repository [here](https://github.com/kurtosis-tech/geth-lighthouse-package).

If you'd prefer not to run it on your local machine, try it out on the [Kurtosis playground](https://gitpod.io/?autoStart=true&editor=code#https://github.com/kurtosis-tech/geth-lighthouse-package)
:::

## Introduction 
A testnet is an incredibly valuable tool for any web3 developer, no matter if you’re building a dApp or working on protocol-level changes. It comes as no surprise then that Ethereum has multiple public testnets in addition to a plethora of tools for local development networks (e.g. Ganache, Hardhat, Foundry Anvil).

However, there are cases where an engineer  may need to develop or test functionality that: modifies the protocol itself (execution or consensus layers), necessitates a certain scale, or interacts with another blockchain entirely (e.g. L2s/rollups, bridges, or multi-chain relayers). In these cases, a fully functioning private testnet is required - one where the user has full control over every aspect of the network and its ancillary services. 

:::note
We will review the details on when and how a full, private testnet can be useful in another article.
:::

This guide will walk you through how to build your very own fully functioning, private Ethereum testnet using [`eth-kurtosis`](https://github.com/kurtosis-tech/eth-kurtosis/tree/main) components. In fact, the artifact you’ll end up with at the end of this tutorial will be a special type of environment definition that works at any scale you desire, is completely reproducible for CI workflows, and is modular - meaning you can add or remove other services to your network as you wish.

**What you will do:**

1. Create a local Kurtosis package template
2. Import required dependencies
3. Define how your private Ethereum testnet should look like. This example will leverage the Lighthouse CL client and Geth EL client to build a single, full node over Docker with [`eth-kurtosis`](https://github.com/kurtosis-tech/eth-kurtosis/tree/main).
4. Launch the private testnet locally over Docker
5. Learn about some advanced workflows you can do with `eth-kurtosis`.

**What you will need beforehand to get started:**

- Install [Docker & ensure its running](../get-started/installing-the-cli.md#i-install--start-docker)
- Install [Kurtosis](../get-started/installing-the-cli.md) (or [upgrade to latest](./upgrading-the-cli.md) if you already have it)
- A Github account to leverage the [template repository](https://github.com/kurtosis-tech/package-template-repo)

### 1. Set up an empty Kurtosis package
To begin, create and `cd` into a directory to hold your files:
```bash
mkdir my-testnet && cd my-testnet 
```
Next, create a file called `network_params.yaml` in that folder with the following contents:
```yaml
preregistered_validator_keys_mnemonic: 'giant issue aisle success illegal bike spike
  question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy
  very lucky have athlete'
num_validator_keys_per_node: 64
network_id: '3151908'
deposit_contract_address: '0x4242424242424242424242424242424242424242'
seconds_per_slot: 12
genesis_delay: 10
capella_fork_epoch: 2
deneb_fork_epoch: 500

```
The contents above will be used to define the specific parameters with which to start the network with.

Now, create a `kurtosis.yml` file in the same folder with the following:
```yml
name: github.com/foo-bar/my-testnet
```
Awesome. You have just created the beginnings of your first [Kurtosis package](../advanced-concepts/packages.md)! This package will form the backbone of the environment definition you will use to instantiate and deploy your private testnet. A Kurtosis package is completely reproducible, modular, and will work locally (Docker, Minikube, k3s, etc) or in the cloud, on backends like EC2 or Kubernetes. 

### 2. Import dependencies
Now that you have a local project to house your definition and some parameters to start the network with, its time to actually build the network. First, create a Starlark file called `main.star` and add the following three lines:
```python
import yaml

# main.star
geth = import_module("github.com/kurtosis-tech/geth-package/lib/geth.star")
lighthouse = import_module("github.com/kurtosis-tech/lighthouse-package/lib/lighthouse.star")
with open("./network_params.yaml") as stream:
    network_params = yaml.safe_load(stream)
```

In the first two lines, you're using [Locators](../advanced-concepts/locators.md) to import in `geth.star` and `lighthouse.star` files from Github, making them available to use in your testnet definition. These files themselves are environment definitions that can be used to bootstrap and start up a Geth execution layer client and a Lighthouse consensus layer client as part of your testnet - which is exactly what you will do next.

:::note
Feel free to check out the [`geth.star`](https://github.com/kurtosis-tech/geth-package/blob/main/lib/geth.star) and [`lighthouse.star`](https://github.com/kurtosis-tech/lighthouse-package/blob/main/lib/lighthouse.star) to understand how they work. At a high level, the definition instructs Kurtosis to generate genesis data, set up pre-funded accounts, and then launches the client using the client container images.
:::

Finally, we are converting the local `network_params.yaml` file into a format that can be used in your environment definition using [`yaml.safe_load()`](https://pyyaml.org/wiki/PyYAMLDocumentation#Loader). This is then saved into a variable called `network_params` which you will use later on.

### 3. Define how your testnet gets built
Now that you have all the necessary dependencies, you can start writing the function that will instantiate the network. Within your `main.star` file, add the following 3 new lines:

```python
def run(plan):
    final_genesis_timestamp = geth.generate_genesis_timestamp()
    el_genesis_data = geth.generate_el_genesis_data(plan, final_genesis_timestamp, network_params)
```

What you've just done here is define a function using `run(plan)` to house all of the methods you will use for instantiating the network. Within this method, you will call the [`generate_genesis_timestamp()` function](https://github.com/kurtosis-tech/geth-package/blob/main/lib/geth.star#L58), from the `geth.star` you imported earlier, to generate an abitrary timestamp for the genesis of your network. This is important for time-based forks that you may want to use later on. Next, you will generate some genesis data for the execution layer using [`generate_el_genesis_data`](https://github.com/kurtosis-tech/geth-package/blob/main/lib/geth.star#L43) which was imported from `geth.star` as well. Under the hood, the genesis data is being generated using the Ethereum Foundation's [`ethereum-genesis-generator`](https://github.com/ethpandaops/ethereum-genesis-generator).

You can already see the benefit of composable environment definitions: you don't need to deal with nor understand how the genesis data is being generated. You can rely on the framework built and used by the Ethereum Foundation for your testnet's genesis data.

With some execution layer genesis data in hand, you will now bootstrap the node! Go ahead and add the next 3 lines to your `main.star` file inside the same `def run(plan)` function so that your final result looks like:

```python
# main.star
import yaml

geth = import_module("github.com/kurtosis-tech/geth-package/lib/geth.star")
lighthouse = import_module("github.com/kurtosis-tech/lighthouse-package/lib/lighthouse.star")
with open("./network_params.yaml") as stream:
    network_params = yaml.safe_load(stream)

def run(plan):
    # Generate genesis, note EL and the CL needs the same timestamp to ensure that timestamp based forking works
    final_genesis_timestamp = geth.generate_genesis_timestamp()
    el_genesis_data = geth.generate_el_genesis_data(plan, final_genesis_timestamp, network_params)

    # NEW LINES TO ADD:
    # Run the nodes
    el_context = geth.run(plan, network_params, el_genesis_data)
    lighthouse.run(plan, network_params, el_genesis_data, final_genesis_timestamp, el_context)

    return
```

Here, the Geth client is launched  using the `run()` function in `geth.star` and then returns all the relevant information about the client to `el_context`, including the [Ethereum Node Record](https://github.com/sigp/enr). This information, alongside the network parameters, genesis data, and the genesis timestamp, are then passed in as arguments in the next command: `lighthouse.run()` which bootstraps the Lighthouse consensus layer client.

And that is it! In these short few lines, you now have an environment definition that spins up a full stacking Ethereum node with Geth and Lighthouse over Docker on your local machine.

### 4. Run your new testnet!
Finally, time to give it a spin! Go back to your terminal and from within the `my-testnet` directory, run:
```
kurtosis run .
```

Kurtosis will interpret the environment definition you just wrote, validate that everything will work, and then execute the instructions to instantiate your Ethereum node inside an [enclave](../advanced-concepts/enclaves.md), which is just a sandbox environment that will house your node. Kurtosis will handle the importing of the `lighthouse.star` and `geth.star` files from Github. The output you'll get at the end should look like this:

```bash
Starlark code successfully run. No output was returned.
INFO[2023-08-04T16:07:28+02:00] ==========================================================
INFO[2023-08-04T16:07:28+02:00] ||          Created enclave: tranquil-woodland          ||
INFO[2023-08-04T16:07:28+02:00] ==========================================================
Name:            tranquil-woodland
UUID:            ac3877184757
Status:          RUNNING
Creation Time:   Fri, 04 Aug 2023 16:06:57 CEST

========================================= Files Artifacts =========================================
UUID           Name
8a1de99b7224   1-lighthouse-eth-0-63
271f6e53a7e1   cl-genesis-data
6c116cfcc7d1   el-genesis-data
7549ddc4135a   genesis-generation-config-cl
d266370395ef   genesis-generation-config-el
d204de12687e   geth-prefunded-keys
a069f55dc147   prysm-password

========================================== User Services ==========================================
UUID           Name                                             Ports                                         Status
cb04101e98fd   cl-client-0                                      http: 4000/tcp -> http://127.0.0.1:50646      RUNNING
                                                                metrics: 5054/tcp -> http://127.0.0.1:50647
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:50648
                                                                udp-discovery: 9000/udp -> 127.0.0.1:59240
f377be0f55f8   cl-client-0-validator                            http: 5042/tcp -> 127.0.0.1:50649             RUNNING
                                                                metrics: 5064/tcp -> http://127.0.0.1:50650
19b325f68893   el-client-0                                      engine-rpc: 8551/tcp -> 127.0.0.1:50639       RUNNING
                                                                rpc: 8545/tcp -> 127.0.0.1:50641
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:50640
                                                                udp-discovery: 30303/udp -> 127.0.0.1:49442
                                                                ws: 8546/tcp -> 127.0.0.1:50642
a9608eaf4942   prelaunch-data-generator-cl-genesis-data         <none>                                        RUNNING
a1e33f5b7141   prelaunch-data-generator-cl-validator-keystore   <none>                                        STOPPED
971aaffb412d   prelaunch-data-generator-el-genesis-data         <none>                                        RUNNING
``` 

You'll now see in the `User Services` section all the ports that you will use to connect to and interact with your local node, including the RPC URL. Your port numbers may differ from the one above.

Congratulations! You now have a full Ethereum staking node for all your private testnet needs.

### 5. Advanced Workflows
You may already know what you want to do with the private testnet you've just spun up, and that's great! We hope this was helpful in getting you started and to show you just how easy it was to write your own testnet definition using Kurtosis.

Otherwise, we've got some neat ideas for what you can do next. If you need a hand with any of the below, feel free to let us know in our [Github Discussions](https://github.com/kurtosis-tech/kurtosis/discussions/new/choose) where we and members of our community can help!
* Deploy your node in a Kubernetes cluster for collaborative work and scale it out to multiple nodes! Check out our docs for how to do so [here](https://docs.kurtosis.com/k8s/). 
* Simulate MEV workflows by importing the [MEV Package](https://github.com/kurtosis-tech/mev-package) into your testnet definition. The MEV package deploys and configures the Flashbots suite of products to your local Ethereum testnet and includes: MEV-Boost, MEV-Flood, and MEV-relay, and any required dependencies (postgres & redis). Here's a full example of this set up [here](https://github.com/kurtosis-tech/geth-lighthouse-mev-package).
* Connect other infrastructure (oracles, relayers, etc) to the network by adding more to your `main.star` file! Remember, this is an environment definition and you can import any pre-existing packages that you may find useful. Here are a [few examples](https://github.com/kurtosis-tech/awesome-kurtosis/tree/main)
* Deploy your dApp onto the local network! Hardhat can be used to do so by using the given RPC URL & the `network_id` defined in the `network_params.yaml` you wrote at the beginning. In your case, the `network_id` should be: `3151908`. A more thorough example of this workflow can also be found [here](./how-to-local-eth-testnet.md).

We're currently building out more components of [`eth-kurtosis`](https://github.com/kurtosis-tech/eth-kurtosis/tree/main), which serves as an index of plug-and-play components for Ethereum private testnets. We're building support for more clients - so let us know if there's something you would love to see added to the index! 

### Conclusion

To recap, in this guide you:
* Created a working directory locally for your Kurtosis package
* Wrote a very short environment definition, `main.star`, which imported the client launchers for your node, generated the necessary starting state, and then launched them!

You also saw first-hand how the composability aspect of Kurtosis environment definitions were used to abstract away a lot of the complexities that come with bootstrapping your own node. And because this is entirely reproducible, your team can use this as a private blockchain for validating and testing changes for your application.

We hope this guide was helpful and we'd love to hear from you. Please don't hesitate to share with us what went well, and what didn't, using [`kurtosis feedback`](../cli-reference/feedback.md) from the CLI to file an issue in our [Github](https://github.com/kurtosis-tech/eth-kurtosis/issues) or post your question in our [Github Discussions](https://github.com/kurtosis-tech/kurtosis/discussions).

Thank you!

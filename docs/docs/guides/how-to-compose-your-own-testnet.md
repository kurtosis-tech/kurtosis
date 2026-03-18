---
title: How to build your own private Ethereum testnet with ethereum-package
sidebar_label: Build your own testnet from scratch
slug: /how-to-compose-your-own-testnet
toc_max_heading_level: 2
sidebar_position: 11
---

:::tip
If you prefer to dive into code examples, visit the repository [here](https://github.com/ethpandaops/ethereum-package/tree/main/examples).
:::

## Introduction 
A testnet is an incredibly valuable tool for any web3 developer, no matter if you’re building a dApp or working on protocol-level changes. It comes as no surprise then that Ethereum has multiple public testnets in addition to a plethora of tools for local development networks (e.g. Ganache, Hardhat, Foundry Anvil).

However, there are cases where an engineer  may need to develop or test functionality that: modifies the protocol itself (execution or consensus layers), necessitates a certain scale, or interacts with another blockchain entirely (e.g. L2s/rollups, bridges, or multi-chain relayers). In these cases, a fully functioning private testnet is required - one where the user has full control over every aspect of the network and its ancillary services. 

The most robust and up-to-date way to build a customizable Ethereum testnet is by using the official [ethereum-package](https://github.com/ethpandaops/ethereum-package).

:::note
We will review the details on when and how a full, private testnet can be useful in another article.
:::

**What you will do:**

1. Create a configuration file for ethereum-package
2. Run the network with Kurtosis
3. Verify the deployment
4. Stop the deployment and clean up

### 1. Create a configuration file
We first create a `network_params.yaml` file in which we will list the combination of execution and consensus clients which will participate in the network alongside the id of the network. The file should contain the following configuration:

```yaml
participants:
  - el_type: reth
    cl_type: lighthouse
  - el_type: geth
    cl_type: teku
network_params:
  network_id: "12345"
```

The parameters are defined in the [ethereum-package](https://github.com/ethpandaops/ethereum-package) and many more things can be customized. See section on [Advanced Customization](#advanced-customization).

### 2. Run the network with Kurtosis
We next use the kurtosis CLI to spin up the custom network using the ethereum-package and the newly created `network_params.yaml` file.

```bash
kurtosis run github.com/ethpandaops/ethereum-package --args-file ./network_params.yaml
```

Below is an example of output after the CLI is done running:

```bash
INFO[2026-03-17T20:33:14-04:00] ======================================================= 
INFO[2026-03-17T20:33:14-04:00] ||          Created enclave: delicate-swamp          || 
INFO[2026-03-17T20:33:14-04:00] ======================================================= 
Name:            delicate-swamp
UUID:            8fbb919823f8
Status:          RUNNING
Creation Time:   Tue, 17 Mar 2026 20:30:50 EDT
Flags:           

========================================= Files Artifacts =========================================
UUID           Name
b6838fd98427   1-lighthouse-reth-0-127
afbc76a8212e   2-teku-geth-128-255
f5e27eb5aed1   el_cl_genesis_data
d1a590ca4f5f   final-genesis-timestamp
fcc7109707b8   genesis-el-cl-env-file
e17d8a189014   genesis_validators_root
9eb6395f8d04   jwt_file
4ee246d08e9e   keymanager_file
7f64fa77db6d   prysm-password
422f95f405c5   validator-ranges

========================================== User Services ==========================================
UUID           Name                                             Ports                                         Status
d47d11982fdf   cl-1-lighthouse-reth                             http: 4000/tcp -> http://127.0.0.1:32781      RUNNING
                                                                metrics: 5054/tcp -> http://127.0.0.1:32782   
                                                                quic-discovery: 9001/udp                      
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:32783    
                                                                udp-discovery: 9000/udp                       
16919e524386   cl-2-teku-geth                                   http: 4000/tcp -> http://127.0.0.1:32784      RUNNING
                                                                metrics: 8008/tcp -> http://127.0.0.1:32785   
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:32786    
                                                                udp-discovery: 9000/udp                       
e874ba5f73e9   el-1-reth-lighthouse                             engine-rpc: 8551/tcp -> 127.0.0.1:32773       RUNNING
                                                                metrics: 9001/tcp -> http://127.0.0.1:32774   
                                                                rpc: 8545/tcp -> 127.0.0.1:32771              
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:32775   
                                                                udp-discovery: 30303/udp                      
                                                                ws: 8546/tcp -> 127.0.0.1:32772               
597d54b8d101   el-2-geth-teku                                   engine-rpc: 8551/tcp -> 127.0.0.1:32778       RUNNING
                                                                metrics: 9001/tcp -> http://127.0.0.1:32779   
                                                                rpc: 8545/tcp -> 127.0.0.1:32776              
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:32780   
                                                                udp-discovery: 30303/udp                      
                                                                ws: 8546/tcp -> 127.0.0.1:32777               
33c4927f65f0   validator-key-generation-cl-validator-keystore   <none>                                        RUNNING
a1d82552a605   vc-1-reth-lighthouse                             metrics: 8080/tcp -> http://127.0.0.1:32787   RUNNING
```

According to the output above, `delicate-swamp` is the name of the enclave (the deployment of the network). Yours will be something different.

### 3. Verify the deployment

The name of the enclave created above is `delicate-swamp`. We can confirm that the enclave is running with:

```bash
kurtosis enclave ls
```

which shows my enclase as `RUNNING` with name `delicate-swamp`:

```bash
UUID           Name             Status     Creation Time
8fbb919823f8   delicate-swamp   RUNNING    Tue, 17 Mar 2026 20:30:50 EDT
```

To further validate that the network is working, we will run a basic JSON-RPC call.

First, we need to know the url of the JSON-RPC endpoint of one of our execution clients.

Run the following command to inspect the enclave:

```bash
kurtosis enclave inspect delicate-swamp
```

The command will produce the following result:

```bash
Name:            delicate-swamp
UUID:            8fbb919823f8
Status:          RUNNING
Creation Time:   Tue, 17 Mar 2026 20:30:50 EDT
Flags:           

========================================= Files Artifacts =========================================
UUID           Name
b6838fd98427   1-lighthouse-reth-0-127
afbc76a8212e   2-teku-geth-128-255
f5e27eb5aed1   el_cl_genesis_data
d1a590ca4f5f   final-genesis-timestamp
fcc7109707b8   genesis-el-cl-env-file
e17d8a189014   genesis_validators_root
9eb6395f8d04   jwt_file
4ee246d08e9e   keymanager_file
7f64fa77db6d   prysm-password
422f95f405c5   validator-ranges

========================================== User Services ==========================================
UUID           Name                                             Ports                                         Status
d47d11982fdf   cl-1-lighthouse-reth                             http: 4000/tcp -> http://127.0.0.1:32781      RUNNING
                                                                metrics: 5054/tcp -> http://127.0.0.1:32782   
                                                                quic-discovery: 9001/udp                      
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:32783    
                                                                udp-discovery: 9000/udp                       
16919e524386   cl-2-teku-geth                                   http: 4000/tcp -> http://127.0.0.1:32784      RUNNING
                                                                metrics: 8008/tcp -> http://127.0.0.1:32785   
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:32786    
                                                                udp-discovery: 9000/udp                       
e874ba5f73e9   el-1-reth-lighthouse                             engine-rpc: 8551/tcp -> 127.0.0.1:32773       RUNNING
                                                                metrics: 9001/tcp -> http://127.0.0.1:32774   
                                                                rpc: 8545/tcp -> 127.0.0.1:32771              
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:32775   
                                                                udp-discovery: 30303/udp                      
                                                                ws: 8546/tcp -> 127.0.0.1:32772               
597d54b8d101   el-2-geth-teku                                   engine-rpc: 8551/tcp -> 127.0.0.1:32778       RUNNING
                                                                metrics: 9001/tcp -> http://127.0.0.1:32779   
                                                                rpc: 8545/tcp -> 127.0.0.1:32776              
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:32780   
                                                                udp-discovery: 30303/udp                      
                                                                ws: 8546/tcp -> 127.0.0.1:32777               
33c4927f65f0   validator-key-generation-cl-validator-keystore   <none>                                        RUNNING
a1d82552a605   vc-1-reth-lighthouse                             metrics: 8080/tcp -> http://127.0.0.1:32787   RUNNING

```

In particular, we notice that `el-1-reth-lighthouse` has the following RPC port forwarded: `rpc: 8545/tcp -> 127.0.0.1:32771`. Thus, the JSON-RPC api is accessible through port `32771` for the Reth execution client. Your specific port value might differ.

We can now run a basic JSON-RPC call to get the current block number:

```bash
curl -X POST http://127.0.0.1:32771 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

Running the query multiple time in a row will produce a block number which increases over time, thus confirming our network is working properly!

### 4. Stopping the deployment and cleaning up
Once we are done we can stop the enclave.

```bash
kurtosis enclave stop delicate-swamp
```

Once stopped, we can finish cleanup by removing the network permanently.

```bash
kurtosis enclave rm delicate-swamp
```

## Advanced Customization
The ethereum-package supports hundreds of configurations, including:
- Setting up mev-boost and relays
- Adding L2 rollups (like OP-Stack)
- Mocking gas prices and block times
- Pre-funding specific wallet addresses

Configuration is accomplished by modifying the `network_params.yaml` file. An extensive list of customization options can be found in the (ethereum-package documentation)[https://github.com/ethpandaops/ethereum-package/tree/main?tab=readme-ov-file#configuration].

In addition, ready-to-go configuration files are available as [examples](https://github.com/ethpandaops/ethereum-package/tree/main/examples).

## Conclusion
To recap, we have:

* Created a `network_params.yml` file to customize the network
* Deployed the network with Kurtosis
* Inspected the network and used this information to interact with it
* Cleaned up our deployment

We’d love to hear from you on what went well for you, what could be improved, or to answer any of your questions. Don’t hesitate to reach out via a post in our [discussions forum on Github](https://github.com/kurtosis-tech/kurtosis/discussions/new?category=q-a) or in our [Discord Server](https://discord.gg/jJFG7XBqcY).

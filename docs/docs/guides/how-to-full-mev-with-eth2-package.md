---
title: How to launch a private Ethereum testnet with Flashbot's MEV Boost implementation of Proposer Builder Separation (PBS)
sidebar_label: Launch a testnet with MEV infra
slug: /how-to-full-mev-with-ethereum-package
toc_max_heading_level: 2
sidebar_position: 12
---

:::tip
Here are some quick short-cuts for folks who would prefer:
* To get going right away: install Kurtosis & Docker and then run: `kurtosis run github.com/kurtosis-tech/ethereum-package '{"mev_type": "full"}'`
* To dive into the code example: visit the repository [here](https://github.com/kurtosis-tech/ethereum-package).
* Not to run this package on their local machine: try it out on the [Kurtosis playground](https://gitpod.io/?autoStart=true&editor=code#https://github.com/kurtosis-tech/ethereum-package)
:::

We're elated to share that the [`ethereum-package`](https://github.com/kurtosis-tech/ethereum-package) now supports the Flashbot's implementation of [Proposer-Builder Separation (PBS)](https://ethereum.org/en/roadmap/pbs/) using [MEV-Boost](https://boost.flashbots.net) protocol.

This milestone marks a huge step forward in the journey towards a full, in-protocol PBS implementation for Proof-of-Stake Ethereum as developers across the ecosystem now have a way to instantiate fully functioning testnets to validate functionality, behvaior, and scales across all client combinations with a Builder API implementation (Flashbots', in this case).

Keep reading to learn [how it all works](#brief-overview-of-the-architecture) & [how to get started with the `ethereum-package`](#quickstart).

#### Why `ethereum-package`?
As a reminder, the [`ethereum-package`](https://github.com/kurtosis-tech/ethereum-package) is a reproducible and portable environment definition that should be used to bootstrap & deploy private testnets. The package will function the exact same way locally or in the cloud over Docker or Kubernetes, supports all major Execution Layer (EL) and Consensus Layer (CL) client implementations, and can be scaled to whatever size your team needs - limited only by your underlying hardware/backend.

#### What if I only want the MEV parts?
And if that wasn't enough, Kurtosis environment definitions (known as [Packages](https://docs.kurtosis.com/advanced-concepts/packages/)) are entirely composable, meaning you can define and build-your-own private testnet using only the parts you need and with the option of adding your own services (e.g. MEV searcher tools). Feel free to check out the following [code example](https://github.com/kurtosis-tech/2-el-cl-mev-package/blob/main/main.star).

## Brief overview of the architecture
Explicitly, the [`ethereum-package`](https://github.com/kurtosis-tech/ethereum-package) supports two modes: `full-mev` and `mock-mev`. 

The former mode is valuable for validating behavior between the protocol and out-of-protocol middle-ware infrastructure (e.g. searchers, relayer) and instantiates [`mev-boost`](https://github.com/flashbots/mev-boost), [`mev-relay`](https://github.com/flashbots/mev-boost-relay), [`mev-flood`](https://github.com/flashbots/mev-flood) and Flashbot's Geth-based block builder called [`mev-builder`](https://github.com/flashbots/builder). The latter mode will only spin up [`mev-boost`](https://github.com/flashbots/mev-boost) and a [`mock-builder`](https://github.com/marioevz/mock-builder), which is useful for testing in-protocol behavior like testing if clients are able to call the relayer for a payload via `mev-boost`, reject invalid payloads, or trigger the [circuit breaker](https://hackmd.io/@ralexstokes/BJn9N6Thc) to ensure functionality of the beacon chain.

The `ethereum-package` with MEV emulation are already in use by client teams to help uncover bugs (examples [here](https://github.com/prysmaticlabs/prysm/pull/12736) and [here](https://github.com/NethermindEth/nethermind/commit/4d805769159dc0717aa1ba38cc3ebc53f9a375cf)). 

Everything you see below in the architecture diagram gets configured, initialized, and bootstrapped together by Kurtosis.

![mev-arch](/img/guides/full-mev-infra-arch-diagram.png)

#### Caveats:
* The `mev-boost-relay` service requires Capella at an epoch of non-zero. For the ethereum-package, the Capella fork is set to happen after the first epoch to be started up and fully connected to the CL client.
* Validators (64 per node by default, so 128 in the example in this guide) will get registered with the relay automatically after the 2nd epoch. This registration process is simply a configuration addition to the mev-boost config - which Kurtosis will automatically take care of as part of the set up. This means that the `mev-relay` infrastructure only becomes aware of the existence of the validators after the 2nd epoch.
* After the 3rd epoch, the `mev-relay` service will begin to receive execution payloads (`eth_sendPayload`, which does not contain transaction content) from the `mev-builder` service (or `mock-builder` in `mock-mev` mode).
* Validators will then start to receive validated execution payload headers from the `mev-relay` service (via `mev-boost`) after the 4th epoch. The validator selects the most valuable header, signs the payload, and returns the signed header to the relay - effectively proposing the payload of transactions to be included in the soon-to-be-proposed block. Once the relay verifies the block proposer's signature, the relay will respond with the full execution payload body (incl. the transaction contents) for the validator to use when proposing a `SignedBeaconBlock` to the network.
* You may notice in the `mev-flood` logs that there may be transactions that fail to get processed by the node(s) in your devnet with the following errors: `Error: replacement fee too low [ See: https://links.ethers.org/v5-errors-REPLACEMENT_UNDERPRICED ]`. Don't be alarmed: this can happen when transactions are sent too quickly to the network, resulting in the node receiving transactions with the same nonce. When this happens, the node rejects the transactions becauase the node assumes you're trying to replace the old pending transaction with a new one. You can change the frequency using `mev-flood`'s `--secondsPerBundle (-p )` flag in the [`spam` command](https://github.com/flashbots/mev-flood#send-swaps).

:::note
Quick aside on what `mev-flood` does:
Once the network is online, `mev-flood` will deploy UniV2 smart contracts, provision liquidity on UniV2 pairs, & begin to send a constant stream of UniV2 swap transactions to the network's public mempool. Depending on the mode you're running, either the `mock-builder` or Flashbot's `mev-builder`, the transactions will be bundled into payloads for downstream use by the relayer or by validators themselves. It is important to note that `mev-flood` will only be initialized with the `full-mev` set up and will send transactions with a non-zero block value. Read more about [`mev-flood` here](https://github.com/flashbots/mev-flood). 
:::

## Quickstart
Leveraging the [`ethereum-package`](https://github.com/kurtosis-tech/ethereum-package) is simple. In this short quickstart, you will:
1. Install Docker & Kurtosis locally.
2. Configure your network using a `.yaml` file.
3. Run a single command to launch your network with `full MEV`.
4. Visit the website to witness payloads being delivered from the Relayer to the `mev-boost` sidecar connected to each validator (for block proposals).

#### Install dependencies
* [Install Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal to see all your Docker images.
* [Install Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

#### Configure your network
Next, create a file titled: `ethereum-package-params.yaml` in your working directory and populate it with:
```yaml
participants:
	- el_client_type: geth
		el_client_image: ethereum/client-go:latest
		el_client_log_level: ''
		el_extra_params: []
		cl_client_type: lighthouse
		cl_client_image: sigp/lighthouse:latest
		cl_client_log_level: ''
		beacon_extra_params: []
		validator_extra_params: []
		builder_network_params: null
network_params:
  network_id: '3151908'
  deposit_contract_address: '0x4242424242424242424242424242424242424242'
  seconds_per_slot: 12
  slots_per_epoch: 32
  num_validator_keys_per_node: 64
  preregistered_validator_keys_mnemonic: 'giant issue aisle success illegal bike spike
    question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy
    very lucky have athlete'
  deneb_for_epoch: 500
verifications_epoch_limit: 5
global_client_log_level: info
mev_type: full
```
You will use the above file by passing it in at runtime, effectively enabling you to define the way your network should look using parameters.

#### Launch the network with `full MEV`
Great! You're now ready to bring up your own network. Simply run:
```bash
kurtosis run --enclave eth-network github.com/kurtosis-tech/ethereum-package "$(cat ~/ethereum-package-params.yaml)"
```
Kurtosis will then begin to spin up your private Ethereum testnet with `full MEV`. You will see a stream of text get printed in your terminal as Kurtosis begins to generate genesis files, configure the Ethereum nodes, launch a Grafana and Prometheus instance, and bootstrap the network together with the full suite of MEV products from Flashbots. In ~2 minutes, you should see the following output at the end:
```bash
Starlark code successfully run. Output was:
{
	"grafana_info": {
		"dashboard_path": "/d/QdTOwy-nz/ethereum-merge-kurtosis-module-dashboard?orgId=1",
		"password": "admin",
		"user": "admin"
	}
}

INFO[2023-08-03T11:16:00+02:00] ====================================================
INFO[2023-08-03T11:16:00+02:00] ||          Created enclave: eth-network          ||
INFO[2023-08-03T11:16:00+02:00] ====================================================
Name:            eth-network
UUID:            1d467f353496
Status:          RUNNING
Creation Time:   Thu, 03 Aug 2023 11:06:50 CEST

========================================= Files Artifacts =========================================
UUID           Name
004cb2a16def   1-lighthouse-geth-0-63
e98eee4d8a99   2-lighthouse-geth-64-127
601b49f6e437   cl-forkmon-config
21192db4c9b4   cl-genesis-data
fcdd39be227b   el-forkmon-config
38905cf9e831   el-genesis-data
0ba35b186c20   genesis-generation-config-cl
b477313c48f4   genesis-generation-config-el
b119fb95bd44   geth-prefunded-keys
c4fd103c5447   grafana-config
122cfb453ebe   grafana-dashboards
b86556fccf74   prometheus-config
2d2d99849ff0   prysm-password

========================================== User Services ==========================================
UUID           Name                                       Ports                                                  Status
1bde5712f965   cl-1-lighthouse-geth                       http: 4000/tcp -> http://127.0.0.1:62873               RUNNING
                                                          metrics: 5054/tcp -> http://127.0.0.1:62874
                                                          tcp-discovery: 9000/tcp -> 127.0.0.1:62875
                                                          udp-discovery: 9000/udp -> 127.0.0.1:53993
57f94044300c   cl-1-lighthouse-geth-validator             http: 5042/tcp -> 127.0.0.1:62876                      RUNNING
                                                          metrics: 5064/tcp -> http://127.0.0.1:62877
ae2d5b824656   cl-2-lighthouse-geth                       http: 4000/tcp -> http://127.0.0.1:62879               RUNNING
                                                          metrics: 5054/tcp -> http://127.0.0.1:62880
                                                          tcp-discovery: 9000/tcp -> 127.0.0.1:62878
                                                          udp-discovery: 9000/udp -> 127.0.0.1:65058
c1eb34a91b7e   cl-2-lighthouse-geth-validator             http: 5042/tcp -> 127.0.0.1:62882                      RUNNING
                                                          metrics: 5064/tcp -> http://127.0.0.1:62881
65e1ae6652c9   cl-forkmon                                 http: 80/tcp -> http://127.0.0.1:62933                 RUNNING
e7f673086384   el-1-geth-lighthouse                       engine-rpc: 8551/tcp -> 127.0.0.1:62861                RUNNING
                                                          rpc: 8545/tcp -> 127.0.0.1:62863
                                                          tcp-discovery: 30303/tcp -> 127.0.0.1:62862
                                                          udp-discovery: 30303/udp -> 127.0.0.1:56858
                                                          ws: 8546/tcp -> 127.0.0.1:62860
3048d9aafc12   el-2-geth-lighthouse                       engine-rpc: 8551/tcp -> 127.0.0.1:62866                RUNNING
                                                          rpc: 8545/tcp -> 127.0.0.1:62864
                                                          tcp-discovery: 30303/tcp -> 127.0.0.1:62867
                                                          udp-discovery: 30303/udp -> 127.0.0.1:62287
                                                          ws: 8546/tcp -> 127.0.0.1:62865
70e19424c664   el-forkmon                                 http: 8080/tcp -> http://127.0.0.1:62934               RUNNING
f4bebfdc819b   grafana                                    http: 3000/tcp -> http://127.0.0.1:62937               RUNNING
219954ae8f7e   mev-boost-0                                api: 18550/tcp -> 127.0.0.1:62931                      RUNNING
287c555090c6   mev-boost-1                                api: 18550/tcp -> 127.0.0.1:62932                      RUNNING
2fedae36a1f8   mev-flood                                  <none>                                                 RUNNING
bb163cad3912   mev-relay-api                              api: 9062/tcp -> 127.0.0.1:62929                       RUNNING
5591d1d17ec5   mev-relay-housekeeper                      <none>                                                 RUNNING
026f3744c98f   mev-relay-website                          api: 9060/tcp -> 127.0.0.1:62930                       RUNNING
096fdfb33909   postgres                                   postgresql: 5432/tcp -> postgresql://127.0.0.1:62928   RUNNING
f283002d8c77   prelaunch-data-generator-cl-genesis-data   <none>                                                 RUNNING
af010cabc0ce   prelaunch-data-generator-el-genesis-data   <none>                                                 RUNNING
3fcd033e1d38   prometheus                                 http: 9090/tcp -> http://127.0.0.1:62936               RUNNING
145b673410f9   redis                                      client: 6379/tcp -> 127.0.0.1:62927                    RUNNING
2172a3e173f8   testnet-verifier                           <none>                                                 RUNNING
f833b940ae5b   transaction-spammer                        <none>                                                 RUNNING
```

As you can see above, there is *a lot* going on in your enclave - but don't worry, let's go through everything together.

The first section that gets printed contains some basic metadata about the enclave that was spun up. This includes the name of the enclave `eth-network`, its [Resource Idenfitier](https://docs.kurtosis.com/advanced-concepts/resource-identifier/), your enclave's status, and the time it was created.

Next, you'll see a section dedicated to [Files Artifacts](https://docs.kurtosis.com/advanced-concepts/files-artifacts/), which are Kurtosis' first-class representation of data inside your enclave, stored as compressed TGZ files. You'll notice there are configuration files for the nodes, grafana, and prometheus as well as private keys for pre-funded accounts and genesis-related data. These files artifacts were generated and used by Kurtosis to start the network and abstracts away the complexities and overhead that come with generating validator keys and getting genesis and node config files produced and mounted to the right containers yourself.

Lastly, there is a section called `User Services` which display the number of services (running in Docker containers) that make up your network. You will notice that there are 2 Ethereum nodes, each with a `MEV-Boost` instance spun up & connected to it. In addition to this, you will see the rest of the Flashbots MEV infrastructure including the `mev-relay` suite of services (read more about the `mev-relay` services [here](https://github.com/flashbots/mev-boost-relay/blob/main/ARCHITECTURE.md)) and `mev-flood`. By default, the `ethereum-package` also comes with supporting services which include a fork monitor, redis, postgres, grafana, prometheus, a transaction spammer, a testnet-verifier, and the services used to generate genesis data. Both of the Redis and Postgres instances are required for `mev-relay` to function properly. Each of these services are running in Docker containers inside your local enclave & Kurtosis has automatically mapped each container port to your machine's ephemeral ports for seamless interaction with the services running in your enclave.

#### Visit the website to see registered validators and delivered payloads 
Now that your network is online, you can visit the relay website using the local port mapped to that endpoint. For this example, it will be `127.0.0.1:62930`, but it will be different for you.

![flashbots-website](/img/guides/full-mev-flashbots-website.png)

The screenshot above is what the website looks like after the 4th epoch. You can see that all 128 validators (2 nodes, each with 64 validators) are registered. The table below will display recently delivered and verified payloads from `mev-relay` to the `mev-boost` sidecar on each node.

And there you have it! You've now spun up a private Ethereum testnet over Docker with the Flashbot's implementation of PBS! 

## Roadmap
The inclusion of a Proposer Builder Separation (PBS) implemention was in support of the Ethereum Foundation's efforts to validate functionality and behavior in end-to-end testing (between in-protocol and out-of-protocol infrastructure), as well as the functionality of the beacon chain for in-protocol code paths (e.g. can clients: call for payloads reject invalid payloads, and trigger the circuit breaker when necessary).

The next immediate thing we hope to do is to *decompose* the environment definition into smaller pieces, enabling developers to build-their-own MEV-enabled systems by simply importing only the parts of the MEV infrastructure that they need. We've begun working on this already with [eth-kurtosis](https://github.com/kurtosis-tech/eth-kurtosis), which contains an index of composable building blocks to define your own testnet.

If there are other use cases you had in mind (e.g. fuzzing the network at the protocol level) or have questions about `eth-kurtosis` or this `ethereum-package`, please don't hesitate to reach out!

## Conclusion
This guide was meant to be quick - we hope it was. To recap, you:
- Installed Kurtosis and Docker
- Created a `.yaml` file that contains the necessary parameters for the network
- Ran a single command to spin up a private Ethereum testnet with MEV infrastructure

The `ethereum-package` is available for anyone to use, will work the same way on your local machine or in the cloud, and will run on Docker or Kubernetes.

You saw first-hand how packages, effectively environment definitions, are written once and then can be used by anyone in a very trivial way to reproduce the environment. This accelerates developer velocity by enabling engineers to spend less time on configuring and setting up development and testing frameworks, and more time instead on building the unique features and capabilities for their projects.

Additionally, we hope you also enjoyed the parameterizability aspect of Kurtosis Packages. By changing the `ethereum-package-params.yaml`, you can get a fine-tune your testnet however you see fit. 

We hope this guide was helpful and we'd love to hear from you. Please don't hesitate to share with us what went well, and what didn't, using [`kurtosis feedback`](../cli-reference/feedback.md) from the CLI to file an issue in our [Github](https://github.com/kurtosis-tech/eth-kurtosis/issues) or post your question in our [Github Discussions](https://github.com/kurtosis-tech/kurtosis/discussions).

Thank you!

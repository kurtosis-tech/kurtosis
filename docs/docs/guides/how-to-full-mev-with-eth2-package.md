---
title: How to launch a private Ethereum testnet with in-protocol Proposer Builder Seperation (PBS) emulation
sidebar_label: Launch a testnet with MEV infra
slug: /how-to-full-mev-with-eth2-package
toc_max_heading_level: 2
---

We're elated to share that Ethereum testnets spun up using Kurtosis' [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) now support in-protocol [Proposer-Builder Separation (PBS)](https://ethereum.org/en/roadmap/pbs/) emulation using Flashbot's open-source [MEV-Boost](https://boost.flashbots.net) implementation. This milestone marks a huge step forward in the journey towards a full, in-protocol PBS implementation for Proof-of-Stake Ethereum and is exciting because engineers can now use the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) to instantiate fully functioning testnets to validate functionality, behvaior, and scales across all client combinations *with MEV infrastructure.* Keep reading to learn [how it all works](#architecture--details) & [how to get started with the `eth2-package`](#quickstart).

#### Why `eth2-package`?
As a reminder, the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) is a reproducible and portable environment definition that should be used to bootstrap & deploy private testnets. The package will function the exact same way locally or in the cloud over Docker or Kubernetes, supports all major Execution Layer (EL) and Consensus Layer (CL) client implementations, and can be scaled to whatever size you need - limited only by your underlying hardware/backend.

#### What if I only want the MEV parts?
And if that wasn't enough, Kurtosis environment definitions (known as [Packages](https://docs.kurtosis.com/concepts-reference/packages/)) are entirely composable, meaning you can define and build-your-own private testnet using only the parts you need and with the option of adding your own services (e.g. MEV searcher tools). Feel free to check out [eth-kurtosis](https://github.com/kurtosis-tech/eth-kurtosis) for how to do this!

## Brief overview of the architecture
Explicitly, the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) supports two MEV modes: `full-mev` and `mock-mev`. 

The former mode is valuable for validating behavior between the protocol and out-of-protocol middle-ware infrastructure (e.g. searchers, relayer) and instantiates [mev-boost](https://github.com/flashbots/mev-boost), [mev-relay](https://github.com/flashbots/mev-boost-relay), [mev-flood](https://github.com/flashbots/mev-flood) and Flashbot's Geth-based block builder called [mev-builder](https://github.com/flashbots/builder). The latter mode will only spin up [mev-boost](https://github.com/flashbots/mev-boost) and a [mock-builder](https://github.com/marioevz/mock-builder), which is useful for testing in-protocol behavior like testing if clients are able to call the relayer for a payload via `mev-boost`, reject invalid payloads, or trigger the circuit breaker to ensure functionality of the beacon chain.

![mev-arch](./assets/mev-infra-arch-diagram.png)

## Quickstart
Using the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) is very straight forward. In this short quickstart, you will:
1. Install required dependencies
2. Configure your network
3. Launch your network with `full MEV`
4. Visit the website to witness payloads being delivered.

#### Install dependencies
* [Install Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal.
* [Install Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

#### Configure your network
Next, create a file titled: `eth2-package-params.json` in your working directory and populate it with:
```json
{
    //  Specification of the participants in the network
    "participants": [
        // Each "participants" block represents a single Ethereum full node. 
        // Therefore, filling in 2 participant blocks will spin up 2 nodes
        {
            "el_client_type": "geth",
            "el_client_image": "ethereum/client-go:latest",
            "el_client_log_level": "",
            "el_extra_params": [],
            "cl_client_type": "lighthouse",
            "cl_client_image": "lighthouse: sigp/lighthouse:latest",
            "cl_client_log_level": "",
            "beacon_extra_params": [],
            "validator_extra_params": [],
            "builder_network_params": null
        }
    ],
    "network_params": {
        "network_id": "3151908",
        "deposit_contract_address": "0x4242424242424242424242424242424242424242",
        "seconds_per_slot": 12,
        "slots_per_epoch": 32,
        "num_validator_keys_per_node": 64,
        "preregistered_validator_keys_mnemonic": "giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete",
         "deneb_for_epoch": 500,

    },
    // True by default such that in addition to the Ethereum network:
    //  - A transaction spammer is launched to fake transactions sent to the network
    //  - Forkmon will be launched after CL genesis has happened
    //  - A prometheus will be started, coupled with grafana
    "launch_additional_services": true,

    //  If set, the package will block until a finalized epoch has occurred.
    "wait_for_finalization": false,

    //  If set to true, the package will block until all verifications have passed
    "wait_for_verifications": false,

    //  If set, after the merge, this will be the maximum number of epochs wait for the verifications to succeed.
    "verifications_epoch_limit": 5,

    //  The global log level that all clients should log at
    //  Valid values are "error", "warn", "info", "debug", and "trace"
    //  This value will be overridden by participant-specific values
    "global_client_log_level": "info",

    // Supports three valeus
    // Default: None - no mev boost, mev builder, mev flood or relays are spun up
    // mock - mock-builder & mev-boost are spun up
    // full - mev-boost, relays, flooder and builder are all spun up
    "mev_type": full
}
```
You will use the above file by passing it in at runtime, effectively enabling you to define the way your network should look using parameters.

#### Launch the network with `full MEV`
You can now launch the network 

```bash
Starlark code successfully run. Output was:
{
	"grafana_info": {
		"dashboard_path": "/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1",
		"password": "admin",
		"user": "admin"
	}
}

INFO[2023-08-09T11:16:00+02:00] ====================================================
INFO[2023-08-09T11:16:00+02:00] ||          Created enclave: timid-brook          ||
INFO[2023-08-09T11:16:00+02:00] ====================================================
Name:            timid-brook
UUID:            1d467f353496
Status:          RUNNING
Creation Time:   Wed, 09 Aug 2023 11:06:50 CEST

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






#### Visit the website to see registered validators and delivered payloads 

## Roadmap

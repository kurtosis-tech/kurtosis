# How To: Instantiate a private Ethereum testnet with Proposer-Builder Separation (PBS) emulation

We're elated to share that Ethereum testnets spun up using Kurtosis' [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) now support in-protocol [Proposer-Builder Separation (PBS)](https://ethereum.org/en/roadmap/pbs/) emulation using Flashbot's open-source [MEV-Boost](https://boost.flashbots.net) implementation. This milestone marks a huge step forward in the journey towards a full, in-protocol PBS implementation for Proof-of-Stake Ethereum and is exciting because engineers can now use the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) to instantiate fully functioning testnets to validate functionality, behvaior, and scales across all client combinations *with MEV infrastructure.* Keep reading to learn [how it all works](#architecture--details) & [how to get started with the `eth2-package`](#quickstart).

#### Why `eth2-package`?
As a reminder, the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) is a reproducible and portable environment definition that should be used to bootstrap & deploy private testnets. The package will function the exact same way locally or in the cloud over Docker or Kubernetes, supports all major Execution Layer (EL) and Consensus Layer (CL) client implementations, and can be scaled to whatever size you need - limited only by your underlying hardware/backend.

#### What if I only want the MEV parts?
And if that wasn't enough, Kurtosis environment definitions (known as [Packages](https://docs.kurtosis.com/concepts-reference/packages/)) are entirely composable, meaning you can define and build-your-own private testnet using only the parts you need and with the option of adding your own services (e.g. MEV searcher tools). Feel free to check out [eth-kurtosis](https://github.com/kurtosis-tech/eth-kurtosis) for how to do this!

## Brief overview of the architecture
Explicitly, the [`eth2-package`](https://github.com/kurtosis-tech/eth2-package) supports two MEV modes: `full-mev` and `mock-mev`. 

The former mode is valuable for validating behavior between the protocol and out-of-protocol middle-ware infrastructure (e.g. searchers, relayer) and instantiates [mev-boost](https://github.com/flashbots/mev-boost), [mev-relay](https://github.com/flashbots/mev-boost-relay), [mev-flood](https://github.com/flashbots/mev-flood) and Flashbot's Geth-based block builder called [mev-builder](https://github.com/flashbots/builder). The latter mode will only spin up [mev-boost](https://github.com/flashbots/mev-boost) and a [mock-builder](https://github.com/marioevz/mock-builder), which is useful for testing in-protocol behavior, like how public mempool transactions or searcher bundles get built into specific payload structures.

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

You will use the above file by passing it in at runtime, enabling you to paramterize the entire 
#### Launch the network with `full MEV`

#### 

## Roadmap
---
title: How to set up an n-node Cassandra environment
sidebar_label: Setting up an n-node Cassandra cluster
slug: /how-to-parameterize-cassandra
toc_max_heading_level: 2
sidebar_position: 5
---

Introduction
------------
In this guide, you will set up a 3-node Cassandra cluster in Docker and parameterize the environment definition so it can easily be modified for use in different tests that require an _n_-node Cassandra cluster. Then we will show you how to run remotely hosted packages authored by others, and go through how to package and publish our work to Github for others to use as well.

Specifically, you're going to configure your test environments with a way that allows you to both:
1. Parameterize the environment so another developer using the environment can specify how many nodes they’d like for their system to have, and 
2. Make the environment definition composable so that your environments can be included in tests with other services for different scenarios & use cases.

**Without Kurtosis**

One way to accomplish the above would be to write shell scripts over docker, or over binaries running on bare metal. In this case, we’d have to build these capabilities into our scripts and handle cross-system issues (laptop vs CI) ourselves.

**With Kurtosis**

In this guide, we’re going to use Kurtosis. Kurtosis has composability, parameterizability, and portability built into its environment definition system and runtime. With Kurtosis we can ensure that these environments are runnable on your own laptop or in your favorite CI provider.

Setup
-----
Before you proceed, make sure you have:
* [Installed and started the Docker engine on your local machine][starting-docker]
* [Installed the Kurtosis CLI (or upgraded it to the latest release, if you already have the CLI installed)][installing-the-cli]

:::tip Use the Starlark VS Code Extension
Feel free to use the [official Kurtosis Starlark VS Code extension][vscode-plugin] when writing Starlark with VSCode for features like syntax highlighting, method signature suggestions, hover preview for functions, and auto-completion for Kurtosis custom types.
:::

Instantiate a 3-node Cassandra cluster
--------------------------------------
First, create and `cd` into a directory to hold the project you'll be working on:

```bash
mkdir kurtosis-cass-cluster && cd kurtosis-cass-cluster
```

Next, create a [Starlark][starlark] file called `main.star` inside your new directory with the following contents:

```python
DEFAULT_NUMBER_OF_NODES = 3
CASSANDRA_NODE_PREFIX= "cassandra-node-"
CASSANDRA_NODE_IMAGE = "cassandra:4.0"


CLUSTER_COMM_PORT_ID = "cluster"
CLUSTER_COM_PORT_NUMBER =  7000
CLUSTER_COM_PORT_PROTOCOL = "TCP"


CLIENT_COMM_PORT_ID = "client"
CLIENT_COM_PORT_NUMBER = 9042
CLIENT_COM_PORT_PROTOCOL = "TCP"


FIRST_NODE_INDEX = 0
  
def run(plan, args):
    num_nodes = DEFAULT_NUMBER_OF_NODES

    # Simple check to verify that we have at least 1 node in our cluster
    if num_nodes == 0:
       fail("Need at least 1 node to Start Cassandra cluster got 0")

    # Iteratively add each node to the cluster, with the given names and serviceConfig specified below
    for node in range(0, num_nodes):
       node_name = get_service_name(node)
       config = get_service_config(num_nodes)
       plan.add_service(name = node_name, config = config)
    
    node_tool_check = "nodetool status | grep UN | wc -l | tr -d '\n'"

    check_nodes_are_up = ExecRecipe(
       command = ["/bin/sh", "-c", node_tool_check],
    )

    # Wait for the nodes to be up and ready to establish connections and receive traffic 
    plan.wait(
        service_name = get_first_node_name(),
        recipe = check_nodes_are_up, 
        field = "output", 
        assertion = "==", 
        target_value = str(num_nodes), 
        timeout = "8m", 
        )

    return {"node_names": [get_service_name(x) for x in range(num_nodes)]}

def get_service_name(node_idx):
    return CASSANDRA_NODE_PREFIX + str(node_idx)

def get_service_config(num_nodes):
    seeds = ["cassandra-node-"+str(x) for x in range(0, num_nodes)]
    return ServiceConfig(
        image = CASSANDRA_NODE_IMAGE,
        ports = {
            CLUSTER_COMM_PORT_ID : PortSpec(number = CLUSTER_COM_PORT_NUMBER, transport_protocol = CLUSTER_COM_PORT_PROTOCOL),
            CLIENT_COMM_PORT_ID : PortSpec(number = CLIENT_COM_PORT_NUMBER, transport_protocol = CLIENT_COM_PORT_PROTOCOL),
        },
        env_vars = {
            "CASSANDRA_SEEDS":",".join(seeds),
            # without this set Cassandra tries to take 8G and OOMs
            "MAX_HEAP_SIZE": "512M",
            "HEAP_NEWSIZE": "1M",
        }
    )

def get_first_node_name():
    return get_service_name(FIRST_NODE_INDEX)
```

Finally, save your newly created file and, from the working directory you created, run the following command:

```bash
kurtosis run --enclave cassandra-cluster main.star
```

:::info
Kurtosis will run validation checks against your code to ensure that it will work before spinning up the containers for our 3-node Cassandra cluster. We won’t dive into the details of how validation checks are used by Kurtosis in this guide, but you can read more about them [here][multi-phase-runs].
:::

Your output will look something like:

```bash
INFO[2023-03-28T17:44:20-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-28T17:44:24-03:00] Enclave 'cassandra-cluster' created successfully

> add_service name="cassandra-node-0" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-0' added with service UUID 'ec084228aa2b4e63aea84c10b9c37963'

> add_service name="cassandra-node-1" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-1' added with service UUID 'f605cff291ef495f884c43f9ee9a980c'

> add_service name="cassandra-node-2" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-2' added with service UUID '4bcf767e82f546e3acfa597510efb0e5'

> wait recipe=ExecRecipe(command=["/bin/sh", "-c", "nodetool status | grep UN | wc -l | tr -d '\n'"]) field="output" assertion="==" target_value="3" timeout="8m"
Wait took 33 tries (52.05875544s in total). Assertion passed with following:
Command returned with exit code '0' and the following output: 3

Starlark code successfully run. Output was:
{
	"node_names": [
		"cassandra-node-0",
		"cassandra-node-1",
		"cassandra-node-2"
	]
}
Name:            cassandra-cluster
UUID:            c8027468561c
Status:          RUNNING
Creation Time:   Tue, 28 Mar 2023 16:29:54 -03

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID           Name               Ports                                  Status
ec084228aa2b   cassandra-node-0   client: 9042/tcp -> 127.0.0.1:52503    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:52502
f605cff291ef   cassandra-node-1   client: 9042/tcp -> 127.0.0.1:52508    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:52507
4bcf767e82f5   cassandra-node-2   client: 9042/tcp -> 127.0.0.1:52513    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:52512
```


Congratulations! You’ve used Kurtosis to spin up a three-node Cassandra cluster over Docker. 

### Review
In this section, you created a [Starlark][starlark] file with instructions for Kurtosis to do the following:
1. Spin up 3 Cassandra containers (one for each node), 
2. Bootstrap each node to the cluster,
3. Map the default Cassandra node container ports to ephemeral local machine ports (described in their respective `ServiceConfig`), and
4. Verify using `nodetool` that the cluster is up, running, and has 3 nodes (just as we specified) before returning the names of our nodes.

You now have a simple environment definition for Kurtosis to spin up a 3-node Cassandra cluster. You may now be wondering: but what if I need to change the number of nodes?

Fortunately, Kurtosis environment definitions can be parameterized. We’ll see just how easy it is to do so in the next section.

Parameterize your Cassandra cluster
----------------------------------

Kurtosis enables users to parameterize environment definitions out-of-the-box. If you need to run tests against multiple different configurations of your environment, you'll be to do so without needing maintain different Bash scripts or `docker-compose.yml` files per test.

You can parameterize our Cassandra cluster environment definition by adding 2 lines of code to your `main.star` Starlark file. 

Let’s add in those extra lines now. 

In your `main.star` file, add the following lines in the code block that describes the `plan` object:

```python
def run(plan, args):
	# Default number of Cassandra nodes in our cluster
	num_nodes = 3

	### <----------- NEW CODE -----------> ###
	if "num_nodes" in args:
		num_nodes = args["num_nodes"]
	### <----------- NEW CODE -----------> ###
	
	for node in range(0, num_nodes):
		node_name = get_service_name(node)
		config = get_service_config(num_nodes)
		plan.add_service(
            name = node_name,
			config = config,
        )

	# ...
```

Now save your newly edited `main.star` file and run the following command to blow away your old enclave and spin up a new one with 5 nodes:

```bash
kurtosis clean -a && kurtosis run --enclave cassandra-cluster main.star '{"num_nodes": 5}'
```

It may take a while, but you should now see the following result:

```bash
INFO[2023-03-28T21:45:46-03:00] Cleaning enclaves...
INFO[2023-03-28T21:45:47-03:00] Successfully removed the following enclaves:
e4c49d41cb0f4c54b9e36ff9c0cba18d	cassandra-cluster
INFO[2023-03-28T21:45:47-03:00] Successfully cleaned enclaves
INFO[2023-03-28T21:45:47-03:00] Cleaning old Kurtosis engine containers...
INFO[2023-03-28T21:45:47-03:00] Successfully cleaned old Kurtosis engine containers
INFO[2023-03-28T21:45:47-03:00] Creating a new enclave for Starlark to run inside...
INFO[2023-03-28T21:45:51-03:00] Enclave 'cassandra-cluster' created successfully

> add_service name="cassandra-node-0" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2,cassandra-node-3,cassandra-node-4", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-0' added with service UUID '9a9213d1d84645bf8c76d179e6b2cade'

> add_service name="cassandra-node-1" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2,cassandra-node-3,cassandra-node-4", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-1' added with service UUID '66c6907c6a8a495b9496eaa37c1de42a'

> add_service name="cassandra-node-2" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2,cassandra-node-3,cassandra-node-4", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-2' added with service UUID 'fe9be7991edd40f891e7c2d7f3e14456'

> add_service name="cassandra-node-3" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2,cassandra-node-3,cassandra-node-4", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-3' added with service UUID 'daff154657ce46378928749312275edf'

> add_service name="cassandra-node-4" config=ServiceConfig(image="cassandra:4.0", ports={"client": PortSpec(number=9042, transport_protocol="TCP"), "cluster": PortSpec(number=7000, transport_protocol="TCP")}, env_vars={"CASSANDRA_SEEDS": "cassandra-node-0,cassandra-node-1,cassandra-node-2,cassandra-node-3,cassandra-node-4", "HEAP_NEWSIZE": "1M", "MAX_HEAP_SIZE": "512M"})
Service 'cassandra-node-4' added with service UUID '8ecd6c4b75a64764a87f2ce9d23cd8f0'

> wait recipe=ExecRecipe(command=["/bin/sh", "-c", "nodetool status | grep UN | wc -l | tr -d '\n'"]) field="output" assertion="==" target_value="5" timeout="15m"
Wait took 33 tries (52.686225608s in total). Assertion passed with following:
Command returned with exit code '0' and the following output: 5

Starlark code successfully run. Output was:
{
	"node_names": [
		"cassandra-node-0",
		"cassandra-node-1",
		"cassandra-node-2",
		"cassandra-node-3",
		"cassandra-node-4"
	]
}
INFO[2023-03-28T21:46:55-03:00] ==========================================================
INFO[2023-03-28T21:46:55-03:00] ||          Created enclave: cassandra-cluster          ||
INFO[2023-03-28T21:46:55-03:00] ==========================================================
Name:            cassandra-cluster
UUID:            c7985e32b076
Status:          RUNNING
Creation Time:   Tue, 28 Mar 2023 21:45:47 -03

========================================= Files Artifacts =========================================
UUID   Name

========================================== User Services ==========================================
UUID           Name               Ports                                  Status
9a9213d1d846   cassandra-node-0   client: 9042/tcp -> 127.0.0.1:54740    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:54741
66c6907c6a8a   cassandra-node-1   client: 9042/tcp -> 127.0.0.1:54746    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:54745
fe9be7991edd   cassandra-node-2   client: 9042/tcp -> 127.0.0.1:54761    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:54760
daff154657ce   cassandra-node-3   client: 9042/tcp -> 127.0.0.1:54768    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:54769
8ecd6c4b75a6   cassandra-node-4   client: 9042/tcp -> 127.0.0.1:54773    RUNNING
                                  cluster: 7000/tcp -> 127.0.0.1:54774
```

Congratulations! You’ve just parameterized your Cassandra cluster environment definition and used it to instantiate a 5-node Cassandra cluster. You can run the same command with 2 nodes, or 4 nodes, and it will just work. **Kurtosis environment definitions are completely reproducible and fully parametrizable.**

:::caution
Depending on how many nodes you wish to spin up, the max heap size of each node may cumulatively consume more memory on your local machine than you have available, causing the Starlark script to time-out. Modifying the `MAX_HEAP_SIZE` property in the `ServiceConfig` for the Cassandra nodes may help, depending on your needs.
:::

### Review
How did that work?

The plan object contains all enclave-modifying methods like `add_service`, `remove_service`, `upload_files`, etc. To use arguments, you accessed them via the second parameter of the `run` function, like so:

```python
def run(plan, args):
	...
```

…which is what you used in your `main.star` file originally!

What you did next was add an `if statement` with the `hasattr()` function to tell Kurtosis that if an argument is passed in at runtime by a user, then override the default `num_nodes` variable, which we set as 3, with the user-specified value.

In our case, you passed in:

```bash
'{"num_nodes": 5}'
```

which told Kurtosis to run your `main.star` file with 5 nodes instead of the default of 3 that we began with. 

:::tip
Note that to pass parameters to the run(plan,args) function, a JSON object needs to be passed in as the 2nd position argument after the script or package path:
```bash
kurtosis run my_package.star '{"arg": "my_name"}'
```
:::

In this section, we showed how to use Kurtosis to parameterize an environment definition. Next, we’ll walk through another property of Kurtosis environments: composability.

Making and using composable environment definitions
---------------------------------------------------
You’ve now written an environment definition using Starlark to instantiate a 3-node Cassandra cluster over Docker, and then modified it slightly to parameterize that definition for other use cases (n-node Cassandra cluster).

However, the benefits of parametrized and reproducible environment properties are amplified when you’re able to share your definition with others to use, or when you use definitions that others (friends, colleagues, etc) have already written.

To quickly demonstrate the latter capability, run:

```bash
kurtosis clean -a && kurtosis run --enclave cassandra-cluster github.com/kurtosis-tech/cassandra-package '{"num_nodes": 5}' 
```

which should give you the same result you got when you ran your local `main.star` file with 5 nodes specified as an argument. However this time, you’re actually using a `main.star` file hosted [remotely on Github][github-cass-package]!

**Any Kurtosis environment definition can be converted into a runnable, shareable artifact that we call a Kurtosis Package.** 

While this guide won’t cover the steps needed to convert your Starlark file and export it for others to use as a Kurtosis Package, you can check out our docs to learn more about how to turn a Starlark file into a [runnable Kurtosis package][runnable-packages].

### Review
You just executed a remote-hosted `main.star` from a Kurtosis package on [Github][github-cass-package]. That remote-hosted `main.star` file has the same code as our local `main.star` file and, with only a [locator][locator], Kurtosis knew that you were referencing a publicly-hosted runnable package. 

The entirety of what you wrote locally in your `main.star` file can be packaged up and pushed to the internet (Github) for anyone to use to instantiate an n-node cassandra cluster by adding a [`kurtosis.yml`][kurtosis-yml] manifest to your directory and publishing the entire directory to Github. Read more about how to do this [here][runnable-packages].

Turning your Starlark code into a runnable Kurtosis package and making it available on Github means any developer will be able to use your Kurtosis package as a building block in their own environment definition or to run different tests using various configurations of nodes in a Cassandra cluster.

This behavior is bi-directional as well. Meaning, you can import any remotely hosted Kurtosis package and use its code inside your own Kurtosis package with the `import_module` instruction like so:

```python
lib = import_module("github.com/foo/bar/src/lib.star")

def run(plan,args)
lib.some_function()
lib.some_variable()
```

**Package distribution via remote-hosted git repositories is one way in which Kurtosis enables environments to be easily composed and connected together without end users needing to know the inner workings of each setup.**

Conclusion
----------
And that’s it! To recap this short guide, you:
1. Wrote an environment definition that instructs Kurtosis to set up a 3 node Cassandra cluster in an isolated environment called an Enclave (over Docker),
2. Added a few lines of code to our `main.star` file to parametrize your environment definition to increase the flexibility and use cases with which it can be used, and,
3. Executed a remotely-hosted Starlark file that was part of what's called a [Kurtosis Package][packages] to demonstrate how your environment definition can be shared with other developers.

We’d love to hear from you on what went well for you, what could be improved, or to answer any of your questions. Don’t hesitate to reach out via Github (`kurtosis feedback`) or in our [Discord server](https://discord.com/channels/783719264308953108/783719264308953111).

### Other exercises you can do with your Cassandra cluster
With your parameterized, reusable environment definition for a multi-node Cassandra cluster, feel free to:
* Turn your local `main.star` file into a runnable Kurtosis package and upload it on Github for others to use following [these instructions][runnable-packages]
* Use our [Go or Typescript SDK][sdk] to write backend integration tests, like this [network partitioning test][network-partitioning-test]

### Other examples
We encourage you to check out our [quickstart][quickstart] (where you’ll build a postgres database and API on top) and our other examples in our [awesome-kurtosis repository][awesome-kurtosis] where you will find other Kurtosis packages for you to check out as well, including a package that spins up a local [Ethereum testnet][eth-package-example] or one that sets up a [voting app using a Redis cluster][redis-package-example]. 

<!---- REFERENCE LINKS BELOW ONLY ---->
[quickstart]: ../quickstart.md
[awesome-kurtosis]: https://github.com/kurtosis-tech/awesome-kurtosis
[multi-phase-runs]: ../concepts-reference/multi-phase-runs.md
[github-cass-package]: https://github.com/kurtosis-tech/cassandra-package/blob/main/main.star
[runnable-packages]: ../concepts-reference/packages.md#runnable-packages
[kurtosis-yml]: ../concepts-reference/kurtosis-yml.md
[locator]: ../concepts-reference/locators.md
[packages]: ../concepts-reference/packages.md
[sdk]: ../client-libs-reference.md
[network-partitioning-test]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/cassandra-network-partition-test
[redis-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/redis-voting-app
[eth-package-example]: https://github.com/kurtosis-tech/eth-network-package
[installing-the-cli]: ./installing-the-cli.md#ii-install-the-cli
[starting-docker]: ./installing-the-cli.md#i-install--start-docker
[starlark]: ../concepts-reference/starlark.md
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension

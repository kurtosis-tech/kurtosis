Advanced Usage
==============

Network Partitioning
--------------------
Kurtosis allows you to write tests that simulate network partitions between different nodes of your network. To use this functionality:

1. Set the `IsPartitioningEnabled` flag to `true` in your `TestConfiguration` object that your test returns
1. Call the `repartition` method on the `NetworkContext` object to divide the network into multiple partitions, configuring the access between the partitions (blocked or not)

This will set the desired network states between the various partitions, simulating a partitioned network. When adding services to the partitioned network, make sure to use `NetworkContext.addServiceToPartition` rather than `addService`, because the latter uses the default partition which will no longer exist after repartitioning.

Mounting External Files
-----------------------
External files can be mounted inside a Docker container running inside a Kurtosis test, so long as the files are packaged inside a GZ-compressed TAR file and hosted at a URL accessible by your CI. To use this functionality:

1. Declare the URLs of the artifacts that your test will need in its `TestConfiguration` object, mapped to IDs that you'll use to identify the artifacts
1. When building the `ContainerCreationConfig` inside your `ContainerConfigFactory`, use the `withFilesArtifacts` function of the builder to declare where to mount the files inside the artifacts

Kurtosis Lambdas
----------------
You might want to provide your testnet definition logic to a wider audience (e.g. your users), or you might want to consume someone else's testnet definition logic. Kurtosis' solution to this are "Lambdas", which are packages of code that execute against the Kurtosis testnet. In practical terms, this means you could write Lambdas that:

* Spin up your system inside of Kurtosis
* Occasionally kill services in the network (like Netflix' [Chaos Monkey](https://netflix.github.io/chaosmonkey/))
* Repartition the network periodically
* Mutate the state of the network

Under the covers, Lambdas are:

1. A gRPC server
1. In an arbitrary language
1. With a connection to the Kurtosis engine 
1. And a single endpoint for executing the Lambda
1. That is packaged inside a Docker image

The containerized nature of Lambdas means that they can be written in your language of choice, and easily distributed via Dockerhub.

The API that a Lambda must implement is defined in [the Lambda API Lib](https://github.com/kurtosis-tech/kurtosis-lambda-api-lib). To see a Lambda example, visit the [Datastore Army Lambda](https://github.com/kurtosis-tech/datastore-army-lambda).

---

[Back to index](https://docs.kurtosistech.com)

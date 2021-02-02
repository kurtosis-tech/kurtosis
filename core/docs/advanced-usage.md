Advanced Usage
==============

Network Partitioning
--------------------
Kurtosis allows you to write tests that simulate network partitions between different nodes of your network. To use this functionality:

1. Set the `IsPartitioningEnabled` flag to `true` in your `TestConfiguration` object that your test returns
1. Call the `repartition` method on the `NetworkContext` object to divide the network into multiple partitions, configuring the access between the partitions (blocked or not)

This will set the desired network states between the various partitions, simulating a partitioned network. When adding services to the partitioned network, make sure to use `NetworkContext.addServiceToPartition` rather than `addService`, because the latter uses the default partition which will no longer exist after repartitioning.

For an example test using this functionality, see [here](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/network_partition_test/network_partition_test_.go).

Mounting External Files
-----------------------
External files can be mounted inside a Docker container running inside a Kurtosis test, so long as the files are packaged inside a GZ-compressed TAR file and hosted at a URL accessible by your CI. To use this functionality:

1. Declare the URLs of the artifacts that your test will need in its `TestConfiguration` object, mapped to IDs that you'll use to identify the artifacts
1. In the `getFilesArtifactMountpoints` function of your `DockerContainerInitializer` implementation, use the artifact IDs to declare where to mount the files inside the artifacts

For an example test using this functionality, see [here](https://github.com/kurtosis-tech/kurtosis-libs/blob/master/golang/testsuite/testsuite_impl/files_artifact_mounting_test/files_artifact_mounting_test_.go).

Further Reading
---------------
* [Quickstart](https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart)
* [Building & Running](./building-and-running.md)
* [Testsuite Customization](./testsuite-customization.md)
* [Debugging common failure scenarios](./debugging-failed-tests.md)
* [Architecture](./architecture.md)
* [Advanced Usage](./advanced-usage.md)
* [Running Kurtosis in CI](./running-in-ci.md)
* [Versioning & upgrading](./versioning-and-upgrading.md)
* [Changelog](./changelog.md) 

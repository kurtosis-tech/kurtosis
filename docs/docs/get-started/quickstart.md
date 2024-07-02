---
title: Quickstart
id: quickstart
sidebar_label: Quickstart
slug: /quickstart
toc_max_heading_level: 2
sidebar_position: 3
---

This guide takes ~5 minutes and will walk you through running a containerized application using the Kurtosis CLI. You'll install Kurtosis, deploy an application from a Github locator, inspect your running environment, and then modify the deployed application by passing in a JSON config.

Install Kurtosis
--------------------
Before you get started, make sure you have:
* [Installed Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal to see all your Docker images.
* [Installed Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

:::tip
This guide will have you writing Kurtosis Starlark. You can optionally install [the VSCode plugin](https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension) to get syntax highlighting, autocomplete, and documentation.
:::

:::tip Have a Docker Compose setup?
Check out this [guide][running-docker-compose] to run your Docker Compose setup with Kurtosis in one line!
:::

Run a basic package from Github
---------------------------------------

Run the following in your favorite terminal:

```console
kurtosis run github.com/kurtosis-tech/basic-service-package --enclave quickstart
```

You should get output that looks like:

![quickstart-default-run.png](/img/home/quickstart-default-run.png)

By running this command, you can see the [basic concepts][basic-concepts] of Kurtosis at work:

1. `github.com/kurtosis-tech/basic-service-package` is the [package][basic-package] you used, and it contains the logic to spin up your application.
2. Your application runs in an [enclave][basic-enclave], which you named `quickstart` via the `--enclave` flag.
3. Your enclave has both services and [files artifacts][basic-files-artifact], which contain the dynamically rendered configuration files of each service.

Inspect your deployed application
--------------------

Command-click, or copy-and-paste to your browser, the URL next to the service called `service-c-1` in your CLI output. This local port binding is handled automatically by Kurtosis, ensuring no port conflicts happen on your local machine as you work with your environments. You should see a simple frontend:

![quickstart-default-service-c-frontend.png](/img/home/quickstart-default-service-c-frontend.png)

Service C depends on Service A and Service B, and has a configuration file containing their private IP addresses that it can use to communicate with them. To check that this is true, copy the files artifact containing this config file out of the enclave:

```console
kurtosis files download quickstart service-c-rendered-config
```

This will put the `service-c-rendered-config` [files artifact][files-artifacts-reference] on your machine. You can see its contents with:

```console
cat service-c-rendered-config/service-config.json
```
You should see the rendered config file with the contents:
```
{
    "service-a": [{"name": "service-a-1", "uri": "172.16.12.4:8501"}],
    "service-b": [{"name": "service-b-1", "uri": "172.16.12.5:8501"}]
}
```

In this step, you saw two ways to interact with your enclave:

1. Accessing URLs via automatically generated local port bindings
2. Transferring [files artifacts][files-artifacts-reference] to your machine, for inspection

<details><summary>More ways to interact with an enclave</summary>

You can also do a set of actions you would expect from a standard Docker or Kubernetes deployments, like:
1. Shell into a service: `kurtosis service shell quickstart service-c-1`
2. See a service's logs: `kurtosis service logs quickstart service-c-1`
3. Execute a command on a service: `kurtosis service exec quickstart service-c-1 'echo hello world'`

</details>

Modify your deployed application with a JSON config
----------

Kurtosis packages take in JSON parameters, allowing developers to make high-level modifications to their deployed applications. To see how this works, run:

```console
kurtosis run --enclave quickstart github.com/kurtosis-tech/basic-service-package \
  '{"service_a_count": 2,
    "service_b_count": 2,
    "service_c_count": 1,
    "party_mode": true}'
```

This runs the same application, but with 2 instances of Service A and Service B (perhaps to test high availability), and a feature flag turned on across all three services called `party_mode`. Your output should look like:

![quickstart-params-output.png](/img/home/quickstart-params-output.png)

If you go to the URL of any of the services, for example Service C, you will see the feature flag `party_mode` is enabled:

![quickstart-service-c-partying.png](/img/home/quickstart-service-c-partying.png)

Each service is partying, but they're each partying for different reasons at the configuration level. By changing the JSON input to the package, you did all of these:
- Changed number of instances of Service A and Service B
- Turned on a feature flag on Service A using its configuration file
- Turned on a feature flag on Service B using a command line flag to its server process
- Turned on a feature flag on Service C using an environment variable on its container

To inspect how each of these changes happened, check out the following:

<details><summary><b>See that the count of each service changed</b></summary>

You can see 2 instances of Service A and 2 instances of Service B in the CLI output:

![quickstart-params-output.png](/img/home/quickstart-params-output.png)

You can verify that the configuration file of Service C has been properly changed so it can talk to all 4 of them:

```console
kurtosis files download quickstart service-c-rendered-config
```
```console
cat service-c-rendered-config/service-config.json
```
You should see the rendered config file with the contents:
```
{
    "service-a": [{"name": "service-a-1", "uri": "172.16.16.4:8501"},{"name": "service-a-2", "uri": "172.16.16.7:8501"}],
    "service-b": [{"name": "service-b-1", "uri": "172.16.16.5:8501"},{"name": "service-b-2", "uri": "172.16.16.8:8501"}]
}
```

</details>

<details><summary><b>See a feature flag turned on by a configuration file on disk</b></summary>

Service A has the `party_mode` flagged turned on by virtue of its configuration file. You can see that with by downloading the `service-a-rendered-config` files artifact, as you've seen before:

```console
kurtosis files download quickstart service-a-rendered-config
```
```console
cat service-a-rendered-config/service-config.json
```
You should see the config file contents with the feature flag turned on:
```
{
    "party_mode": true
}
```

</details>

<details><summary><b>See a feature flag turned on by an command line argument</b></summary>

Service B has the `party_mode` flag turned on by virtue of a command line flag. To see this, run:
```console
kurtosis service inspect quickstart service-b-1
```

You should see, in the output, the CMD block indicating that the flag was passed as a command line argument to the server process:
```console
CMD:
  --
  --party-mode
```

</details>

<details><summary><b>See a feature flag turned on by an environment variable</b></summary>

Service C has the `party_mode` flag turned on by virtue of an environment variable. To see the environment variable flag is indeed enabled, run:

```console
kurtosis service inspect quickstart service-c-1
```

In the output, you will see a block called `ENV:`. In that block, you should see the environment variable `PARTY_MODE: true`.

</details>

With a JSON (or YAML) interface to packages, developers don't have to dig through low-level docs, or track down the maintainers of Service A, B, or C to learn how to deploy their software in each of these different ways. They just use the arguments of the package to get their environments the way they want them.

--------

Now that you've use the Kurtosis CLI to run a package, inspect the resulting environment, and then modify it by passing in a JSON config, you can take any of these next steps:

- To continue working with Kurtosis by using packages that have already been written, take a look through our [code examples][code-examples].
- To learn how to deploy packages over Kubernetes, instead of over your local Docker engine, take a look at our guide for [running Kurtosis over k8s][running-in-k8s]
- To learn how to write your own package, check out our guide on [writing your first package][write-your-first-package].


<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ../get-started/installing-the-cli.md
[installing-docker-guide]: ../get-started/installing-the-cli.md#i-install--start-docker
[upgrading-kurtosis-guide]: ../guides/upgrading-the-cli.md
[basic-concepts]: ../get-started/basic-concepts.md
[basic-enclave]: ../get-started/basic-concepts.md#enclave
[basic-package]: ../get-started/basic-concepts.md#package
[basic-files-artifact]: ../get-started/basic-concepts.md#files-artifact
[write-your-first-package]: ../get-started/write-your-first-package.md
[running-in-k8s]: ../guides/running-in-k8s.md
[running-docker-compose]: ../guides/running-docker-compose.md
[self-cloud-hosting]: ../guides/self-cloud-hosting.md

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
[code-examples]: ../code-examples.md

<!--------------------------- Other ------------------------------------>
<!-- Examples repo -->
[awesome-kurtosis-repo]: https://github.com/kurtosis-tech/awesome-kurtosis
[data-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/dvd-rental-data.tar
[cassandra-package-example]: https://github.com/kurtosis-tech/cassandra-package
[go-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/go-test
[ts-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/ts-test
[ethereum-package]: https://github.com/ethpandaops/ethereum-package/

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

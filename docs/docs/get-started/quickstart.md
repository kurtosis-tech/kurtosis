---
title: Quickstart
id: quickstart
sidebar_label: Quickstart
slug: /quickstart
toc_max_heading_level: 2
sidebar_position: 3
---

This guide takes ~5 minutes and will walk you through running a containerized application using the Kurtosis CLI. You'll also inspect your running environment, modify the deployed application, and then run it over Kubernetes.

<details><summary>Forget installing! Let me do it on Gitpod</summary>

If you don't want to install Docker and Kurtosis, you can run through the quickstart on Gitpod:
 
**Open the Playground: [Start](https://gitpod.io/?autoStart=true&editor=code#https://github.com/kurtosis-tech/ethereum-package)**

Click on the "New Workspace" button! You don't have to worry about the Context URL, Editor or Class. It's all pre-configured for you.
 
</details>


Install Kurtosis
--------------------
Before you get started, make sure you have:
* [Installed Docker](https://docs.docker.com/get-docker/) and ensure the Docker Daemon is running on your machine (e.g. open Docker Desktop). You can quickly check if Docker is running by running: `docker image ls` from your terminal to see all your Docker images.
* [Installed Kurtosis](https://docs.kurtosis.com/install/#ii-install-the-cli) or [upgrade Kurtosis to the latest version](https://docs.kurtosis.com/upgrade). You can check if Kurtosis is running using the command: `kurtosis version`, which will print your current Kurtosis engine version and CLI version.

Run a basic package from Github
---------------------------------------

Run the following in your favorite terminal:

```console
kurtosis run github.com/kurtosis-tech/basic-service-package --enclave quickstart
```

You should get output that looks like:

![quickstart-default-run.png](/img/home/quickstart-default-run.png)

By running this command, you've seen three [basic concepts][basic-concepts] of Kurtosis:

1. The [package][basic-package] you used, remotely hosted on Github at `github.com/kurtosis-tech/basic-service-package`
2. The [enclave][basic-enclave] you created, which was named `quickstart` via the `--enclave` flag.
3. The [files artifacts][files-artifacts-reference] stored in the enclave, which represent the rendered configuration files of each service.

Inspect your deployed application
--------------------

Cmd-click, or copy-and-paste to your browser, the URL next to the service called `service-c-1` in your CLI output. This port binding is handled automatically by Kurtosis, ensuring no port conflicts happen on your local machine as you work with your environments. You should see a simple frontend, looking something like:

![quickstart-default-service-c-frontend.png](/img/home/quickstart-default-service-c-frontend.png)

Here, Service C is claiming to depend on Service A and Service B, and has a configuration file containing the private IP addresses that it can use to communicate to Service A and Service B. To verify this is true, download the files artifact representing this config file with:

```console
kurtosis files download quickstart service-c-rendered-config
```

This will put the `service-c-rendered-config` [files artifact][files-artifacts-reference] on your machine. You can see its contents with:

```console
~ cat service-c-rendered-config/service-config.json
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

Modify your deployed application with a JSON configuration
----------

Kurtosis packages take in JSON parameters, allowing developers to make high-level modifications to their deployed applications without needing to know the lower-level details like which environment variables, command line arguments, or configuration files to change. To see how this works, lets run the same application with 2 instances of Service A and Service B (perhaps to test that high availability is functioning), and a feature flag turned on across all three services called `party_mode`:

```console

```


<!-- !!!!!!!!!!!!!!!!!!!!!!!!!!! ONLY LINKS BELOW HERE !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! -->

<!--------------------------- Guides ------------------------------------>
[installing-kurtosis-guide]: ../get-started/installing-the-cli.md
[installing-docker-guide]: ../get-started/installing-the-cli.md#i-install--start-docker
[upgrading-kurtosis-guide]: ../guides/upgrading-the-cli.md
[basic-concepts]: ../get-started/basic-concepts.md
[basic-enclave]: ../get-started/basic-concepts.md#enclave
[basic-package]: ../get-started/basic-concepts.md#package
[how-to-set-up-postgres-guide]: write-your-first-package.md

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

<!--------------------------- Other ------------------------------------>
<!-- Examples repo -->
[awesome-kurtosis-repo]: https://github.com/kurtosis-tech/awesome-kurtosis
[data-package-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/data-package
[data-package-example-main.star]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/main.star
[data-package-example-seed-tar]: https://github.com/kurtosis-tech/awesome-kurtosis/blob/main/data-package/dvd-rental-data.tar
[cassandra-package-example]: https://github.com/kurtosis-tech/cassandra-package
[go-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/go-test
[ts-test-example]: https://github.com/kurtosis-tech/awesome-kurtosis/tree/main/quickstart/ts-test
[ethereum-package]: https://github.com/kurtosis-tech/ethereum-package/

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

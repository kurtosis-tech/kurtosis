---
title: Running Kurtosis in CI
sidebar_label: Running in CI
slug: /ci
---

Running Kurtosis on your local machine is nice, but executing it as part of CI is even better. This guide will walk you through modifying your CI config file to use Kurtosis in your CI environment:

Step One: Installing The CLI
----------------------------
You'll need the Kurtosis CLI inside your CI environment. This can be accomplished by following [the installation instructions](installing-the-cli.md) for whichever package manager your CI container uses. E.g. if you're using Github Actions with an Ubuntu executor, you'd add the instructions for installing the Kurtosis CLI via the `apt` package manager to your CI config file.

Step Two: Initialize The Configuration
--------------------------------------
When the Kurtosis CLI is executed for the first time on a machine, we ask you to make a choice about whether you'd like to send anonymized usage metrics to help us make the product better (explanation of why we do this, and how we strive to do this ethically, [here](../explanations/metrics-philosophy.md)). CI environments are non-interactive, so this prompt would cause the CLI running in CI to hang until the CI job times out.

To solve this problem, the Kurtosis CLI includes the `config init` subcommand to non-interactively initialize the CLI's configuration. This one-time call will save your election just as if you'd answered the prompt, so that when the CLI is run the prompt won't be displayed.

You'll therefore want the first call to the `kurtosis` CLI in your CI job to be either:

```
kurtosis config init send-metrics
``` 

if you'd like to help us make the product better for you or 

```
kurtosis config init dont-send-metrics
``` 

if you'd prefer not to send metrics.

Step Three: Start The Engine
----------------------------
You'll need the Kurtosis engine to be running to interact with Kurtosis, both via the [CLI](../reference/cli.md) and the [SDK](../reference/sdk.md). Add `kurtosis engine start` in your CI config file after the CLI installation commands so that your Kurtosis commands work.

Step Four: Run Your Custom Logic
---------------------------------
This will be specific to whatever you want to run in CI. E.g. if you have Javascript Mocha tests that use Kurtosis, you'd put that in your CI config file after installing the Kurtosis CLI & starting the engine.

Step Five: Capturing Enclave Output
-----------------------------------
Naturally, if your job fails you'll want to see what was going on inside of Kurtosis at the time of failure. The `kurtosis enclave dump` command allows us to capture container logs & specs from an enclave, so that we can dump the state of the enclaves and attach them to the CI job for further debugging. The specifics of how to attach files to a CI job from within the job will vary depending on which CI provider you're using, but will look something like the following (which is for CircleCI):

```yaml
      # Run our custom logic (in this case, running a package), but don't exit immediately if it fails so that
      # we can upload the 'enclave dump' results before the CI job ends
      - run: |
          if ! kurtosis run --enclave-id my-enclave github.com/kurtosis-tech/datastore-army-package; then
            touch /tmp/testsuite-failed
          fi

      # Dump enclave data so we can debug any issues that arise
      - run: |
          cd /tmp

          # Write enclave information to /tmp/my-enclave
          kurtosis enclave dump my-enclave my-enclave

          # Zip up the data so we can attach it to the CI job
          zip -r my-enclave.zip my-enclave
      
      # Attach the ZIP file to the CI job
      - store_artifacts:
          path: /tmp/my-enclave.zip
          destination: my-enclave.zip

      # Now that we've uploaded the enclave data, fail the job if the testsuite failed
      - run: "! [ -f /tmp/testsuite-failed ]"
```

Example
-------
- [CircleCI](https://github.com/kurtosis-tech/eth2-package/blob/master/.circleci/config.yml#L19)
- More CI examples coming soon...

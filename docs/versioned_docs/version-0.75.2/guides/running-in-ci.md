---
title: Running Kurtosis in CI
sidebar_label: Running in CI
slug: /ci
---

Running Kurtosis on your local machine is nice, but executing it as part of CI is even better. This guide will walk you through modifying your CI config file to use Kurtosis in your CI environment:

I. Installing The CLI
----------------------------
You'll need the Kurtosis CLI inside your CI environment. This can be accomplished by following [the installation instructions](installing-the-cli.md) for whichever package manager your CI container uses. E.g. if you're using Github Actions with an Ubuntu executor, you'd add the instructions for installing the Kurtosis CLI via the `apt` package manager to your CI config file.

II. Start The Engine
----------------------------
You'll need the Kurtosis engine to be running to interact with Kurtosis, both for the [CLI](../cli-reference/index.md) and [using the client libraries](../client-libs-reference.md). Add `kurtosis engine start` in your CI config file after the CLI installation commands so that your Kurtosis commands work.

III. Run Your Custom Logic
---------------------------------
This will be specific to whatever you want to run in CI. E.g. if you have Javascript Mocha tests that use Kurtosis, you'd put that in your CI config file after installing the Kurtosis CLI & starting the engine.

IV. Capturing Enclave Output
-----------------------------------
Naturally, if your job fails you'll want to see what was going on inside of Kurtosis at the time of failure. The `kurtosis enclave dump` command allows us to capture container logs & specs from an enclave, so that we can dump the state of the enclaves and attach them to the CI job for further debugging. The specifics of how to attach files to a CI job from within the job will vary depending on which CI provider you're using, but will look something like the following (which is for CircleCI):

```yaml
      # Run our custom logic (in this case, running a package), but don't exit immediately if it fails so that
      # we can upload the 'enclave dump' results before the CI job ends
      - run: |
          if ! kurtosis run --enclave-name my-enclave github.com/kurtosis-tech/datastore-army-package; then
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
- [CircleCI](https://github.com/kurtosis-tech/eth2-package/blob/main/.circleci/config.yml#L19)
- More CI examples coming soon...

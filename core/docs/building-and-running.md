Building & Running
==================
Every testsuite is simply a package of test code in [an arbitrary language](https://github.com/kurtosis-tech/kurtosis-docs/blob/master/supported-languages.md) that runs in a Docker container. This means that every test developer needs to a) build a testsuite Docker image and b) then feed it to Kurtosis for execution.

### Building a testsuite container
Testsuites bootstrapped using [the quickstart instructions](./quickstart.md) will come with the Dockerfile and `main` function necessary to package your test code into a Docker image. To build the testsuite Docker image, you'd need to call `docker build` on the Dockerfile to generate a new Docker image every time you make changes to your testsuite. This becomes tedious quickly, so we've automated this with a script that we'll see soon.

### Feeding a testsuite container to Kurtosis
Kurtosis is invoked via the `kurtosis.sh` Bash script that's [released with each version of Kurtosis Core](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=wrapper-script/). To run your testsuite, you'd need to call `kurtosis.sh` and pass in the name of your testsuite image. As with building, this becomes tedious so we've automated it in a script.

### build_and_run.sh
Building the testsuite image and running it is such a common task that calling `docker build` and `kurtosis.sh` manually each time becomes frictionful. To ease the pain, we've automated the process with a script called `build_and_run.sh` in the `scripts` directory of every bootstrapped repo. `build_and_run.sh` naturally builds the testsuite Docker image and runs it via `kurtosis.sh`, and it takes an argument instructing it to just build, just run, or both (call `build_and_run.sh help` to see the options available).

Because `build_and_run.sh` will call down to `kurtosis.sh` and `kurtosis.sh` has arguments of its own, any additional arguments after the arg telling `build_and_run.sh` what to do will be passed as-is to `kurtosis.sh`. As an example, you can call `build_and_run.sh all --help` to signify that a) `build_and_run.sh` should do both build and run steps and b) you want to see the extra flags that the inner call to `kurtosis.sh` receives. As a second example, `build_and_run.sh run --parallelism 2` would execute only the run step (no build) and call `kurtosis.sh` with parallelism set to 2.

Next Steps
----------
Now that you understand more about the internals of a testsuite, you can:

* Head over to [the quickstart instructions](./quickstart.md) to bootstrap your own testsuite (if you haven't already)
* Visit [the architecture docs](./architecture.md) to learn more about the Kurtosis platform at a high level.
* Check out [the instructions for running in CI](./running-in-ci.md) to see what's necessary to get Kurtosis running in your CI environment
* Pop into [the Kurtosis Discord](https://discord.gg/6Jjp9c89z9) to join the community!

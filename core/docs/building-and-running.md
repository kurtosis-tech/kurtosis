Building & Running
==================
Every testsuite is simply a package of test code in [an arbitrary language](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/blob/master/supported-languages.txt) that runs in a Docker container. This means that every test developer needs to a) build a testsuite Docker image and b) then feed it to Kurtosis for execution.

### Building a testsuite container
Testsuites bootstrapped using [the quickstart instructions](https://github.com/kurtosis-tech/kurtosis-testsuite-starter-pack/tree/master#testsuite-quickstart) will come with the Dockerfile and `main` function necessary to package your test code into a Docker image. To build the testsuite Docker image, you'd need to call `docker build` on the Dockerfile to generate a new Docker image every time you make changes to your testsuite. This becomes tedious quickly, so we've automated this with a script that we'll see soon.

### Feeding a testsuite container to Kurtosis
Kurtosis is invoked via the `kurtosis.sh` Bash script that's [released with each version of Kurtosis Core after 1.4](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/) and created for you in the `.kurtosis` directory of a bootstrapped testsuite. To run your testsuite, you'd need to call `kurtosis.sh` and pass in the name of your testsuite image. As with building, this becomes tedious so we've automated it in a script.

### build-and-run.sh
Building the testsuite image and running it is such a common task that calling `docker build` and `kurtosis.sh` manually each time becomes frictionful. To ease the pain, we've automated the process with a script called `build-and-run-core.sh`, which is also [released with every Kurtosis Core version](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/) and will also be created in the `.kurtosis` directory of newly-bootstrapped repos alongside `kurtosis.sh`. This script takes in an action (e.g. `build`, to just build, or `help`, to display helptext) as well as several other arguments, and will handle the image-building and image-running for you.

`build-and-run-core.sh` is versioned with Kurtosis Core and should be upgraded whenever Kurtosis Core is upgraded. This means that any changes you make to it would be overwritten. To fix this issue, every bootstrapped testsuite repo comes with a `build-and-run.sh` wrapper script that calls down to `build-and-run-core.sh`, and is yours to modify as you please.

Because `build-and-run.sh` calls `build-and-run-core.sh` which calls `kurtosis.sh`, and because `kurtosis.sh` has arguments of its own, any additional arguments after the arg telling `build-and-run.sh` what to do will be passed as-is to `kurtosis.sh`. As an example, you can call `build-and-run.sh all --help` to signify that a) `build-and-run.sh` should do both build and run steps and b) you want to see the extra flags that the inner call to `kurtosis.sh` receives. As a second example, `build-and-run.sh run --parallelism 2` would execute only the run step (no build) and call `kurtosis.sh` with parallelism set to 2.

---

[Back to index](https://docs.kurtosistech.com)

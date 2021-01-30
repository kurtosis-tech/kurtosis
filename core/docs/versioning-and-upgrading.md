Kurtosis Versioning & Upgrading
===============================
Versioning Scheme
-----------------
Kurtosis follows a modified [Semver](https://semver.org/) where in an `X.Y.Z` version string, _both_ major (`X`) and minor (`Y`) version changes signify API-breaking changes. Patch (`Z`) version changes will not introduce any API-breaking changes. The reason for this modification is because strict Semver often results in silly versions like `238.1.1`. With the major version incremented so high and no higher version to roll over to, there's no way to signify a Really Big Change (like a major refactor). Our method fixes this.

To minimize bugs and keep Kurtosis users running the latest patch version, the kurtosis-core Docker [initializer image](https://hub.docker.com/r/kurtosistech/kurtosis-core_initializer) and [API image](https://hub.docker.com/r/kurtosistech/kurtosis-core_api) will only be tagged `X.Y`. This ensures users will always running the latest patch release.

Upgrading
---------
The version of Kurtosis is encapsulated in the `kurtosis.sh` script that you use to launch it. To upgrade Kurtosis, simply download the latest version of the wrapper script from [the public Kurtosis S3 bucket](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=wrapper-script/) and replace the version you have in your repo.

Further Reading
---------------
* [Quickstart](./quickstart.md)
* [Building & Running](./building-and-running.md)
* [Debugging common failure scenarios](./debugging-failed-tests.md)
* [Architecture](./architecture.md)
* [Advanced Usage](./advanced-usage.md)
* [Running Kurtosis in CI](./running-in-ci.md)
* [Supported languages](./supported-languages.md)
* [Versioning & upgrading](./versioning-and-upgrading.md)
* [Changelog](./changelog.md) 

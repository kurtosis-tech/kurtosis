Kurtosis Versioning & Upgrading
===============================
Versioning Scheme
-----------------
Kurtosis follows a modified [Semver](https://semver.org/) where in an `X.Y.Z` version string, _both_ major (`X`) and minor (`Y`) version changes signify API-breaking changes. Patch (`Z`) version changes will not introduce any API-breaking changes. The reason for this modification is because strict Semver often results in silly versions like `238.1.1`. With the major version incremented so high and no higher version to roll over to, there's no way to signify a Really Big Change (like a major refactor). Our method fixes this.

To minimize bugs and keep Kurtosis users running the latest patch version, the kurtosis-core Docker [initializer image](https://hub.docker.com/r/kurtosistech/kurtosis-core_initializer) and [API image](https://hub.docker.com/r/kurtosistech/kurtosis-core_api) will only be tagged `X.Y`. This ensures users will always running the latest patch release.

Upgrading
---------
### Kurtosis Core
The version of Kurtosis Core is encapsulated in the `kurtosis.sh` script that you use to launch it that lives inside the `.kurtosis` directory in the root of your testsuite. To upgrade Kurtosis Core:

1. Review [the changelog](./changelog.md) for breaking changes between your current Kurtosis Core version and your desired version
1. Replace all files in your `.kurtosis` directory with [the files of the version you're upgrading to](https://kurtosis-public-access.s3.us-east-1.amazonaws.com/index.html?prefix=dist/)
1. Fix any breaks

### Kurtosis Lib
Your testsuite will also depend on a language-specific Kurtosis Lib, which lets your testsuite interact with Kurtosis Core. When upgrading Kurtosis Core, you'll need to upgrade your Kurtosis Lib to match the newly-updated Core version. 

These versions move independently - a single version of Kurtosis Lib might be compatible with multiple versions of Kurtosis Core and vice versa - so we aim to put out an upgrade script that will navigate the compatibility chart transparently for the user. 

Until then, we recommend always upgrading to the latest Kurtosis Core & Lib versions (as the latest versions of both will always be compatible).

---

[Back to index](https://docs.kurtosistech.com)

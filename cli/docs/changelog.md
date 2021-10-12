# TBD

# 0.3.1
### Fixes
* Pinned the default API container version to the same version as in the `go.mod`, so that its version can't silently upgrade under users and break everything

# 0.3.0
### Changes
* Changed the Homebrew/deb/rpm package name to `kurtosis-cli` (was `kurtosis`)

### Breaking Changes
* The CLI is now installed via the `kurtosis-cli` package (for Homebrew, APT, and Yum) rather than just `kurtosis`

# 0.2.1
### Fixes
* Fixed missing `FURY_TOKEN` when publishing

# 0.2.0
### Features
* Ported over the CLI & internal testsuite from `kurtosis-core`

### Breaking Changes
* Changed pretty much everything to add the CLI

# 0.1.0
* Initial commit

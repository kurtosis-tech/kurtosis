# Changelog

## [0.68.2](https://github.com/kurtosis-tech/kurtosis/compare/0.68.1...0.68.2) (2023-03-08)


### Bug Fixes

* fix push_cli_artifacts ci job ([#118](https://github.com/kurtosis-tech/kurtosis/issues/118)) ([b905870](https://github.com/kurtosis-tech/kurtosis/commit/b90587057b200e7f54d1ef5a7e815a1d94a7cf4c))

## [0.68.1](https://github.com/kurtosis-tech/kurtosis/compare/0.68.0...0.68.1) (2023-03-08)


### Features

* docs are versioned ([#106](https://github.com/kurtosis-tech/kurtosis/issues/106)) ([7cd6a4e](https://github.com/kurtosis-tech/kurtosis/commit/7cd6a4e391d7b261cdb2d94d3d9dac2be7f3490b))

## [0.68.0](https://github.com/kurtosis-tech/kurtosis/compare/0.67.4...0.68.0) (2023-03-07)


### ⚠ BREAKING CHANGES

* Migrate Kurtosis Print instruction to Starlark framework. This restrict the use of `print` to a single argument only. ([#80](https://github.com/kurtosis-tech/kurtosis/issues/80)) (#87)

### Features

* enclave clean has both name and uuid ([#101](https://github.com/kurtosis-tech/kurtosis/issues/101)) ([69114ab](https://github.com/kurtosis-tech/kurtosis/commit/69114ab455715092060d51d854f18241f0fb4060))
* persist partition connection overrides to disk ([#98](https://github.com/kurtosis-tech/kurtosis/issues/98)) ([4af3b9f](https://github.com/kurtosis-tech/kurtosis/commit/4af3b9f31daf4962a1e4242a001d2d4bcc84f8d0))


### Code Refactoring

* Migrate Kurtosis Print instruction to Starlark framework. This restrict the use of `print` to a single argument only. ([#80](https://github.com/kurtosis-tech/kurtosis/issues/80)) ([#87](https://github.com/kurtosis-tech/kurtosis/issues/87)) ([868da1b](https://github.com/kurtosis-tech/kurtosis/commit/868da1b871f5b2610dfcc97618c13861a180cc80))

## [0.67.4](https://github.com/kurtosis-tech/kurtosis/compare/0.67.3...0.67.4) (2023-03-04)


### Features

* added new `service_name` parameter for the `exec`, `request` and `wait` instructions. NOTE: the previous methods' signature will be maintained during a deprecation period, we suggest users update the methods' calls to this new signature. ([#66](https://github.com/kurtosis-tech/kurtosis/issues/66)) ([1b47ee3](https://github.com/kurtosis-tech/kurtosis/commit/1b47ee3bb3fd56711995596fb9f68c5a195291fb))
* added the `id` flag in the `analytics` CLI command which allow users to get the `analytics ID` in an easy way ([#81](https://github.com/kurtosis-tech/kurtosis/issues/81)) ([766c094](https://github.com/kurtosis-tech/kurtosis/commit/766c0944a983a0f26e2f7bb3f24ce20f3db28d4b))
* integrate nature theme based name to cli (render template and store service) for file artifacts ([#82](https://github.com/kurtosis-tech/kurtosis/issues/82)) ([aea5bef](https://github.com/kurtosis-tech/kurtosis/commit/aea5bef1fdbd16f88bc4021e243d60f24491b616))
* integrate nature theme named to render_template and store_service ([aea5bef](https://github.com/kurtosis-tech/kurtosis/commit/aea5bef1fdbd16f88bc4021e243d60f24491b616))
* introduce nature themed name for enclaves ([#59](https://github.com/kurtosis-tech/kurtosis/issues/59)) ([78e363f](https://github.com/kurtosis-tech/kurtosis/commit/78e363f554494891b28b4e277e3b04473a66af7b))
* persist service partitions ([#84](https://github.com/kurtosis-tech/kurtosis/issues/84)) ([d46d92a](https://github.com/kurtosis-tech/kurtosis/commit/d46d92a1f0a1db3ba2099e31570983faa0d93874))


### Bug Fixes

* handle multiline errors that might happen with kurtosis clean ([#69](https://github.com/kurtosis-tech/kurtosis/issues/69)) ([f7400be](https://github.com/kurtosis-tech/kurtosis/commit/f7400beac0c7a7f2ec04486064d7bf0c63758cf5))

## [0.67.3](https://github.com/kurtosis-tech/kurtosis/compare/0.67.2...0.67.3) (2023-02-28)


### Features

* Add new FR, docs, and Bug Report issues templates ([#52](https://github.com/kurtosis-tech/kurtosis/issues/52)) ([8854585](https://github.com/kurtosis-tech/kurtosis/commit/88545857213f25716abf4030f03cdd71db531c83))
* made the name field optional for file artifacts in starlark ([#51](https://github.com/kurtosis-tech/kurtosis/issues/51)) ([1ded385](https://github.com/kurtosis-tech/kurtosis/commit/1ded385720423f58a168b44afb94279d1d2c3951))


### Bug Fixes

* Correct minor error in "locators" reference docs ([#71](https://github.com/kurtosis-tech/kurtosis/issues/71)) ([3d68919](https://github.com/kurtosis-tech/kurtosis/commit/3d68919aafbc16e8211cd7692d1820bbe7301070))
* stamp enclave uuid at the end of enclave objects ([#74](https://github.com/kurtosis-tech/kurtosis/issues/74)) ([4f44d03](https://github.com/kurtosis-tech/kurtosis/commit/4f44d03769c877fc36349a79a47b347d2444cf75))

## [0.67.2](https://github.com/kurtosis-tech/kurtosis/compare/0.67.1...0.67.2) (2023-02-27)


### Features

* added boilerplate method to generate unique file artifact name ([#40](https://github.com/kurtosis-tech/kurtosis/issues/40)) ([50cd25c](https://github.com/kurtosis-tech/kurtosis/commit/50cd25cddeccbadf450e7888155b3b39f58acd4b))
* fix the output of kurtosis enclave dump ([#62](https://github.com/kurtosis-tech/kurtosis/issues/62)) ([7ae12cf](https://github.com/kurtosis-tech/kurtosis/commit/7ae12cf51f966a64b3684f3ad439befb8bf2d7c1))


### Bug Fixes

* enforced kurtosis locator validations when running remote kurtosis package ([#41](https://github.com/kurtosis-tech/kurtosis/issues/41)) ([e9af4d9](https://github.com/kurtosis-tech/kurtosis/commit/e9af4d9701e5ecc5b53811d839563140cdc5de22))
* preserve cli provided ordering of completions throughout shells ([#61](https://github.com/kurtosis-tech/kurtosis/issues/61)) ([f312f2c](https://github.com/kurtosis-tech/kurtosis/commit/f312f2c276b335f64c87fd8e34a7fdca5814a75c))

## [0.67.1](https://github.com/kurtosis-tech/kurtosis/compare/0.67.0...0.67.1) (2023-02-23)


### Features

* added Kurtosis Docs command ([#34](https://github.com/kurtosis-tech/kurtosis/issues/34)) ([2502bae](https://github.com/kurtosis-tech/kurtosis/commit/2502baecdfa57dabd8e3bb0d69569c38e6f27645))


### Bug Fixes

* better errors when enclave cleaning fails ([#47](https://github.com/kurtosis-tech/kurtosis/issues/47)) ([a15fe52](https://github.com/kurtosis-tech/kurtosis/commit/a15fe5282652e406e779dfad37fa9ee8cf8ed771))
* enforce kurtosis.yml validations in import_module and read_file; package name inside kurtosis.yml must be valid and is same as the path where kurtosis.yml exists ([#24](https://github.com/kurtosis-tech/kurtosis/issues/24)) ([95d5548](https://github.com/kurtosis-tech/kurtosis/commit/95d554808eaf07928058285016bf6f3a5aff9359))
* fix error message on importing/reading a package instead of a module ([#33](https://github.com/kurtosis-tech/kurtosis/issues/33)) ([1f906ae](https://github.com/kurtosis-tech/kurtosis/commit/1f906ae5dc70a48b670ddda8065e12b81a9bb55c))
* fixed link to report docs issues ([#36](https://github.com/kurtosis-tech/kurtosis/issues/36)) ([dfccf10](https://github.com/kurtosis-tech/kurtosis/commit/dfccf10c01aa5c981fb60fce97725a427fc4d1be))

## [0.67.0](https://github.com/kurtosis-tech/kurtosis/compare/0.66.11...0.67.0) (2023-02-21)


### ⚠ BREAKING CHANGES

* This is a breaking change where we are deprecating PacketDelay to introduce latency in favour of PacketDelayDistribution. Instead of using packet delay, use UniformPacketDelayDistribution for constant delays or NormalPacketDelayDistribution for normally distributed latencies

## [0.66.11](https://github.com/kurtosis-tech/kurtosis/compare/0.66.10...0.66.11) (2023-02-21)


### Features

* track enclave size after run has finished ([#15](https://github.com/kurtosis-tech/kurtosis/issues/15)) ([80f35c8](https://github.com/kurtosis-tech/kurtosis/commit/80f35c80797b00fd66b4d216b9807b63cf2739b7))


### Bug Fixes

* import_module and read_file should load files from kurtosis packages (kurtosis.yml must be present in the path). enforce that only kurtosis packages (directories containing kurtosis.yml) can be run. ([#16](https://github.com/kurtosis-tech/kurtosis/issues/16)) ([84f1042](https://github.com/kurtosis-tech/kurtosis/commit/84f1042aef2d79b388bb5eaf808df520a3e76462))

## [0.66.10](https://github.com/kurtosis-tech/kurtosis/compare/0.66.9...0.66.10) (2023-02-16)


### Features

* made metrics opt in by default ([#5](https://github.com/kurtosis-tech/kurtosis/issues/5)) ([cd076fd](https://github.com/kurtosis-tech/kurtosis/commit/cd076fd515a05e594f338e693405d614718e59f4))
* update metrics lib to track os, arch & backend type ([#11](https://github.com/kurtosis-tech/kurtosis/issues/11)) ([15cf9bb](https://github.com/kurtosis-tech/kurtosis/commit/15cf9bbec3b37a6235901d01e207be36366e8614))

## [0.66.9](https://github.com/kurtosis-tech/kurtosis/compare/0.66.8...0.66.9) (2023-02-15)


### Bug Fixes

* Fix release process ([#3](https://github.com/kurtosis-tech/kurtosis/issues/3)) ([8a618d9](https://github.com/kurtosis-tech/kurtosis/commit/8a618d94bebe0553f744ab90b6e4c8fe2f8f7fdb))

## 0.66.8 (2023-02-15)


### Features

* Initial implementation ([#1](https://github.com/kurtosis-tech/kurtosis/issues/1)) ([8a3fd81](https://github.com/kurtosis-tech/kurtosis/commit/8a3fd8123388de117f4a8c84622024923d410fc3))

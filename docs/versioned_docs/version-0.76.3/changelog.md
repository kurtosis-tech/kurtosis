# Changelog

## [0.76.3](https://github.com/kurtosis-tech/kurtosis/compare/0.76.2...0.76.3) (2023-04-27)


### Bug Fixes

* make api depend not on internal version of grpc-file-transfer ([#572](https://github.com/kurtosis-tech/kurtosis/issues/572)) ([8cb536e](https://github.com/kurtosis-tech/kurtosis/commit/8cb536e35930e11d0414e8ab49a2bfa86692fe2d))

## [0.76.2](https://github.com/kurtosis-tech/kurtosis/compare/0.76.1...0.76.2) (2023-04-27)


### Bug Fixes

* fixed grpc-file-transfer Golang module name ([#570](https://github.com/kurtosis-tech/kurtosis/issues/570)) ([bcb0dc9](https://github.com/kurtosis-tech/kurtosis/commit/bcb0dc935ee8b22c0900d6cdaf844c6e78baab14))

## [0.76.1](https://github.com/kurtosis-tech/kurtosis/compare/0.76.0...0.76.1) (2023-04-26)


### Bug Fixes

* random error message after execution ([#565](https://github.com/kurtosis-tech/kurtosis/issues/565)) ([daedaef](https://github.com/kurtosis-tech/kurtosis/commit/daedaef4a82ad49f0dcdf865c716b72e919d48c5))

## [0.76.0](https://github.com/kurtosis-tech/kurtosis/compare/0.75.9...0.76.0) (2023-04-26)


### ⚠ BREAKING CHANGES

* Added automatic service's ports opening wait for TCP and UDP ports. All the declared service's TCP and UDP ports will be checked by default but this can be also deactivate. This change should not break anything in most cases but could be some cases were the default timeout is not enough and users will have to increase the wait timeout to fix the break  ([#534](https://github.com/kurtosis-tech/kurtosis/issues/534))

### Features

* Added automatic service's ports opening wait for TCP and UDP ports. All the declared service's TCP and UDP ports will be checked by default but this can be also deactivate. This change should not break anything in most cases but could be some cases were the default timeout is not enough and users will have to increase the wait timeout to fix the break  ([#534](https://github.com/kurtosis-tech/kurtosis/issues/534)) ([a961b6e](https://github.com/kurtosis-tech/kurtosis/commit/a961b6e03edc91abad0a12a277bb083062fbe2a0))

## [0.75.9](https://github.com/kurtosis-tech/kurtosis/compare/0.75.8...0.75.9) (2023-04-24)


### Features

* allow passing an exec to shell  ([#550](https://github.com/kurtosis-tech/kurtosis/issues/550)) ([44c9187](https://github.com/kurtosis-tech/kurtosis/commit/44c91876dbee951de70368db33a379237a7f8cda))
* Raise file size limit to 100mb for file downloads and uploads ([#542](https://github.com/kurtosis-tech/kurtosis/issues/542)) ([ec8534a](https://github.com/kurtosis-tech/kurtosis/commit/ec8534aeb187f3c17b69c344d96efe24cc187697))
* replace runtime values in output with real values ([#541](https://github.com/kurtosis-tech/kurtosis/issues/541)) ([8df9666](https://github.com/kurtosis-tech/kurtosis/commit/8df966631afca0fbfe0bd345fe9a0576b55824f6))


### Bug Fixes

* restrict random network allocation to 10.0.0.0/8 following RFC 4096 ([#545](https://github.com/kurtosis-tech/kurtosis/issues/545)) ([003f190](https://github.com/kurtosis-tech/kurtosis/commit/003f190af636f76009fac34899d8b51ef5dad901))

## [0.75.8](https://github.com/kurtosis-tech/kurtosis/compare/0.75.7...0.75.8) (2023-04-20)


### Features

* GRPC file streaming library ([#504](https://github.com/kurtosis-tech/kurtosis/issues/504)) ([ec30ada](https://github.com/kurtosis-tech/kurtosis/commit/ec30ada5e81e18442c60e420c4fb86435a79d2a5))


### Bug Fixes

* added telemetry about network partitioning for enclaves that get created ([#535](https://github.com/kurtosis-tech/kurtosis/issues/535)) ([379a1a6](https://github.com/kurtosis-tech/kurtosis/commit/379a1a69404f04c9f6d6235e1759c471951c0419))

## [0.75.7](https://github.com/kurtosis-tech/kurtosis/compare/0.75.6...0.75.7) (2023-04-19)


### Bug Fixes

* prune non 0 patch versions for docs that aren't current minor version ([#528](https://github.com/kurtosis-tech/kurtosis/issues/528)) ([c65d691](https://github.com/kurtosis-tech/kurtosis/commit/c65d69168fcbcd3c7e470dedb4156594616e35a4)), closes [#487](https://github.com/kurtosis-tech/kurtosis/issues/487)

## [0.75.6](https://github.com/kurtosis-tech/kurtosis/compare/0.75.5...0.75.6) (2023-04-19)


### Features

* validate port ids before execution ([#519](https://github.com/kurtosis-tech/kurtosis/issues/519)) ([f6aceee](https://github.com/kurtosis-tech/kurtosis/commit/f6aceee42f65ce239b019d4179543aaf53b9e605))


### Bug Fixes

* Fix error message in recipe extraction logic to help with debugging ([#527](https://github.com/kurtosis-tech/kurtosis/issues/527)) ([45f9f8b](https://github.com/kurtosis-tech/kurtosis/commit/45f9f8b8d2b01d3480e444bd9319a048966802ca))
* Fix NPE when stopping an already killled Portal process ([#526](https://github.com/kurtosis-tech/kurtosis/issues/526)) ([7307322](https://github.com/kurtosis-tech/kurtosis/commit/7307322bdf36e8dac21cf613c40fbab78e426685))
* Pass Content-Type header to request ([#531](https://github.com/kurtosis-tech/kurtosis/issues/531)) ([b3a9294](https://github.com/kurtosis-tech/kurtosis/commit/b3a92943258493c0bc705c3755d8b9ae20715035))

## [0.75.5](https://github.com/kurtosis-tech/kurtosis/compare/0.75.4...0.75.5) (2023-04-18)


### Features

* Add extractors to exec recipe ([#508](https://github.com/kurtosis-tech/kurtosis/issues/508)) ([b044093](https://github.com/kurtosis-tech/kurtosis/commit/b0440932e18397239212c63576bb15fbda00bd59))

## [0.75.4](https://github.com/kurtosis-tech/kurtosis/compare/0.75.3...0.75.4) (2023-04-18)


### Bug Fixes

* Address flakiness of extractor test ([#510](https://github.com/kurtosis-tech/kurtosis/issues/510)) ([4508df3](https://github.com/kurtosis-tech/kurtosis/commit/4508df328b5f91310353cec3e7abb58483e40083))
* Support ExecRecipe in ReadyCondition ([#507](https://github.com/kurtosis-tech/kurtosis/issues/507)) ([539e8e9](https://github.com/kurtosis-tech/kurtosis/commit/539e8e97185aa785d74c814bf587b06bd9f6ed04))

## [0.75.3](https://github.com/kurtosis-tech/kurtosis/compare/0.75.2...0.75.3) (2023-04-18)


### Bug Fixes

* Re-enable `--platform=linux/amd64` for Engine and APIC docker image ([#516](https://github.com/kurtosis-tech/kurtosis/issues/516)) ([07a0d07](https://github.com/kurtosis-tech/kurtosis/commit/07a0d07250e30fbf422917005e706c1a10789750))

## [0.75.2](https://github.com/kurtosis-tech/kurtosis/compare/0.75.1...0.75.2) (2023-04-17)


### Bug Fixes

* Fix argument extraction logic when argument is present but is of wrong type ([#514](https://github.com/kurtosis-tech/kurtosis/issues/514)) ([0c97d83](https://github.com/kurtosis-tech/kurtosis/commit/0c97d83daea233d1687bc7a56dfd39035c1fc4d3))
* use subnetworking over partitioning ([#491](https://github.com/kurtosis-tech/kurtosis/issues/491)) ([327cdcf](https://github.com/kurtosis-tech/kurtosis/commit/327cdcfb5b6d97805bcd9bf4bbbee7eb2270af94)), closes [#443](https://github.com/kurtosis-tech/kurtosis/issues/443)
* wait command times out even if recipe is still executing ([#480](https://github.com/kurtosis-tech/kurtosis/issues/480)) ([9fd81fa](https://github.com/kurtosis-tech/kurtosis/commit/9fd81fadeb8662c39c20c2647b2fb9e2c5946506))

## [0.75.1](https://github.com/kurtosis-tech/kurtosis/compare/0.75.0...0.75.1) (2023-04-11)


### Bug Fixes

* revert files download enclave flag to arg  ([#489](https://github.com/kurtosis-tech/kurtosis/issues/489)) ([6844393](https://github.com/kurtosis-tech/kurtosis/commit/68443939d27e3eb249ae75eebb913b09877a53b8)), closes [#460](https://github.com/kurtosis-tech/kurtosis/issues/460)

## [0.75.0](https://github.com/kurtosis-tech/kurtosis/compare/0.74.0...0.75.0) (2023-04-10)


### ⚠ BREAKING CHANGES

* removed the Kurtosis CLI `config init` command ([#461](https://github.com/kurtosis-tech/kurtosis/issues/461))

### Code Refactoring

* removed the Kurtosis CLI `config init` command ([#461](https://github.com/kurtosis-tech/kurtosis/issues/461)) ([06578e4](https://github.com/kurtosis-tech/kurtosis/commit/06578e4cad2a097daa8e1dd77c252b97b370606d)), closes [#435](https://github.com/kurtosis-tech/kurtosis/issues/435)

## [0.74.0](https://github.com/kurtosis-tech/kurtosis/compare/0.73.2...0.74.0) (2023-04-03)


### ⚠ BREAKING CHANGES

* renamed the `ReadyConditions` custom type  to `ReadyCondition` ([#429](https://github.com/kurtosis-tech/kurtosis/issues/429))

### Features

* Add linting validation for Markdown reference-style links ([#453](https://github.com/kurtosis-tech/kurtosis/issues/453)) ([7cbe728](https://github.com/kurtosis-tech/kurtosis/commit/7cbe72869c8f3ac86db0f13dea107ad5f54a5dd6)), closes [#448](https://github.com/kurtosis-tech/kurtosis/issues/448)


### Bug Fixes

* colourized the cli output and display starlark messages to the log ([#414](https://github.com/kurtosis-tech/kurtosis/issues/414)) ([af072af](https://github.com/kurtosis-tech/kurtosis/commit/af072af845a887a21171988cb470d29cb70b4884))


### Code Refactoring

* renamed the `ReadyConditions` custom type  to `ReadyCondition` ([#429](https://github.com/kurtosis-tech/kurtosis/issues/429)) ([4076ec4](https://github.com/kurtosis-tech/kurtosis/commit/4076ec4cbc9a04da7ba7060af0e9c5be89f866ff))

## [0.73.2](https://github.com/kurtosis-tech/kurtosis/compare/0.73.1...0.73.2) (2023-04-02)


### Bug Fixes

* Fix two small bugs in the docs ([#451](https://github.com/kurtosis-tech/kurtosis/issues/451)) ([d960dee](https://github.com/kurtosis-tech/kurtosis/commit/d960dee0a4db4a253e766ae04f23f24ab08e985a))

## [0.73.1](https://github.com/kurtosis-tech/kurtosis/compare/0.73.0...0.73.1) (2023-04-02)


### Features

* Reduce the word count & language complexity of the Github Issue templates ([#437](https://github.com/kurtosis-tech/kurtosis/issues/437)) ([b1fad7d](https://github.com/kurtosis-tech/kurtosis/commit/b1fad7d9207be855fbdbce9d70410aecf679d892))


### Bug Fixes

* Actually fix the broken Docusaurus links in the navbar and footer ([#450](https://github.com/kurtosis-tech/kurtosis/issues/450)) ([3436445](https://github.com/kurtosis-tech/kurtosis/commit/3436445f3f6351b66c9ef5d86a59accd19f4baaf))
* Fix bug with release-please PR getting a modified Yarn lockfile ([#446](https://github.com/kurtosis-tech/kurtosis/issues/446)) ([bfa155b](https://github.com/kurtosis-tech/kurtosis/commit/bfa155bf4d4be19cff1f3635083d6390586b94fa))
* Fix links that should be absolute ([#427](https://github.com/kurtosis-tech/kurtosis/issues/427)) ([492bea6](https://github.com/kurtosis-tech/kurtosis/commit/492bea61723b03377c1b981c946fd3fd1c83c70e))
* Fixed many broken links in the docs ([#444](https://github.com/kurtosis-tech/kurtosis/issues/444)) ([9251cc9](https://github.com/kurtosis-tech/kurtosis/commit/9251cc9f49a323c8916112decc9cd9d9e1fcc430))
* Improve error message when package arg fails deserialisation ([#418](https://github.com/kurtosis-tech/kurtosis/issues/418)) ([d54fd73](https://github.com/kurtosis-tech/kurtosis/commit/d54fd73e0cb3713214549d6d20f04d374d8bbb50))
* Indent Caused by in stacktraces ([#417](https://github.com/kurtosis-tech/kurtosis/issues/417)) ([c165a15](https://github.com/kurtosis-tech/kurtosis/commit/c165a15ca24e5af523e27c2d34661025e4189590))
* Remove no-dead-links Remark plugin ([#447](https://github.com/kurtosis-tech/kurtosis/issues/447)) ([b59b3f8](https://github.com/kurtosis-tech/kurtosis/commit/b59b3f8fd07fe3789595b580085841df54990b57))
* Remove Quickstart, SDK, and CLI links from the navbar ([#449](https://github.com/kurtosis-tech/kurtosis/issues/449)) ([a7effc9](https://github.com/kurtosis-tech/kurtosis/commit/a7effc946c5db2eeacbcbaee4286c85989a7005f))
* Try and fix Yarn lockfile violation that's causing Cloudflare publish to fail ([#445](https://github.com/kurtosis-tech/kurtosis/issues/445)) ([4db878b](https://github.com/kurtosis-tech/kurtosis/commit/4db878ba1c47e55af92206249775b573b8de7fd0))

## [0.73.0](https://github.com/kurtosis-tech/kurtosis/compare/0.72.2...0.73.0) (2023-03-31)


### ⚠ BREAKING CHANGES

* Moved the `sevice_name` argument to the first position in the `exec`, `request`, and `wait` instructions, users will have to adapt these instructions calls if where using positional arguments. ([#412](https://github.com/kurtosis-tech/kurtosis/issues/412))

### Features

* Add portal `status`, `start` and `stop` command ([#334](https://github.com/kurtosis-tech/kurtosis/issues/334)) ([beec527](https://github.com/kurtosis-tech/kurtosis/commit/beec5275f3344d81ea4c30553743d7524857ccf5))
* clean the error for starlark output as well ([#413](https://github.com/kurtosis-tech/kurtosis/issues/413)) ([5953a23](https://github.com/kurtosis-tech/kurtosis/commit/5953a23413ec6ee07790e1330dd6f0389e959b6c))


### Bug Fixes

* clean error paths for users ([#369](https://github.com/kurtosis-tech/kurtosis/issues/369)) ([fedc8d0](https://github.com/kurtosis-tech/kurtosis/commit/fedc8d0a82b387498e00f5dabf40c7fbf40247f8))


### Code Refactoring

* Moved the `sevice_name` argument to the first position in the `exec`, `request`, and `wait` instructions, users will have to adapt these instructions calls if where using positional arguments. ([#412](https://github.com/kurtosis-tech/kurtosis/issues/412)) ([126ccbc](https://github.com/kurtosis-tech/kurtosis/commit/126ccbcc5920714af14bc47bc7190867d6279803))

## [0.72.2](https://github.com/kurtosis-tech/kurtosis/compare/0.72.1...0.72.2) (2023-03-30)


### Bug Fixes

* Make GetCluster non fatal if it is unset ([#403](https://github.com/kurtosis-tech/kurtosis/issues/403)) ([3e9db8f](https://github.com/kurtosis-tech/kurtosis/commit/3e9db8f736c8d25f513d080c367f30011d5da511))

## [0.72.1](https://github.com/kurtosis-tech/kurtosis/compare/0.72.0...0.72.1) (2023-03-30)


### Features

* Noop when switching to current context ([#390](https://github.com/kurtosis-tech/kurtosis/issues/390)) ([2003df9](https://github.com/kurtosis-tech/kurtosis/commit/2003df912278fe4fd30e29ab9011ebb39834d7b5))


### Bug Fixes

* Fix confusing warning about port mapping ([#396](https://github.com/kurtosis-tech/kurtosis/issues/396)) ([2bc2e13](https://github.com/kurtosis-tech/kurtosis/commit/2bc2e13de487be3e4458c2ac2c0d000ce0d12589))
* fix help text and have flags instead of args for files download ([#395](https://github.com/kurtosis-tech/kurtosis/issues/395)) ([f9083cc](https://github.com/kurtosis-tech/kurtosis/commit/f9083cc34584dd2face61a7165bdfa2b8be2f0ba)), closes [#370](https://github.com/kurtosis-tech/kurtosis/issues/370)

## [0.72.0](https://github.com/kurtosis-tech/kurtosis/compare/0.71.2...0.72.0) (2023-03-30)


### ⚠ BREAKING CHANGES

* Change starlark args from struct to dict ([#376](https://github.com/kurtosis-tech/kurtosis/issues/376))

### Features

* Restart engine post cluster set ([#393](https://github.com/kurtosis-tech/kurtosis/issues/393)) ([be82680](https://github.com/kurtosis-tech/kurtosis/commit/be82680880552add195954d2962c74e9fecefed0))


### Code Refactoring

* Change starlark args from struct to dict ([#376](https://github.com/kurtosis-tech/kurtosis/issues/376)) ([f350621](https://github.com/kurtosis-tech/kurtosis/commit/f350621f4080514caa96b93a0581186d07a306a6))

## [0.71.2](https://github.com/kurtosis-tech/kurtosis/compare/0.71.1...0.71.2) (2023-03-30)


### Bug Fixes

* Fix engine local dependencies replace ([#389](https://github.com/kurtosis-tech/kurtosis/issues/389)) ([99e4160](https://github.com/kurtosis-tech/kurtosis/commit/99e41605b7f5445b453c5d55aeb2f4d043445666))

## [0.71.1](https://github.com/kurtosis-tech/kurtosis/compare/0.71.0...0.71.1) (2023-03-29)


### Features

* add a timestamp at the end of dump dir if default is chosen ([#386](https://github.com/kurtosis-tech/kurtosis/issues/386)) ([6e1898e](https://github.com/kurtosis-tech/kurtosis/commit/6e1898e999e22ebb1b893b6e65f44d26d059b9d9))
* best effort autocomplete for service logs after engine restart ([#374](https://github.com/kurtosis-tech/kurtosis/issues/374)) ([e2fb47c](https://github.com/kurtosis-tech/kurtosis/commit/e2fb47c927ec673afc63308a7eaa688c0c91bb80)), closes [#373](https://github.com/kurtosis-tech/kurtosis/issues/373)


### Bug Fixes

* polish Github issue templates ([#380](https://github.com/kurtosis-tech/kurtosis/issues/380)) ([bd9a9d0](https://github.com/kurtosis-tech/kurtosis/commit/bd9a9d05afe5e93c792b8dbfe25e84166f175d65))

## [0.71.0](https://github.com/kurtosis-tech/kurtosis/compare/0.70.7...0.71.0) (2023-03-29)


### ⚠ BREAKING CHANGES

* rename the argument `name` to `service_name` for `update_service`, `remove_service` and `add_service` ([#360](https://github.com/kurtosis-tech/kurtosis/issues/360))

### Features

* Automatically map all service ports to local ports post Starlark run and service add ([#363](https://github.com/kurtosis-tech/kurtosis/issues/363)) ([7906aee](https://github.com/kurtosis-tech/kurtosis/commit/7906aee2d3aacb0afcaffb1e77847b9d4148e905))
* rename the argument `name` to `service_name` for `update_service`, `remove_service` and `add_service` ([#360](https://github.com/kurtosis-tech/kurtosis/issues/360)) ([c80d3c0](https://github.com/kurtosis-tech/kurtosis/commit/c80d3c0da7e536590551e5f6c53c9caf4add781c)), closes [#200](https://github.com/kurtosis-tech/kurtosis/issues/200)


### Bug Fixes

* Fixed broken quickstart code block ([#339](https://github.com/kurtosis-tech/kurtosis/issues/339)) ([00f5cd2](https://github.com/kurtosis-tech/kurtosis/commit/00f5cd2576bdf62da2fd071f3cba39f3b976075c))
* Improve error message when cloning a git repo failed ([#375](https://github.com/kurtosis-tech/kurtosis/issues/375)) ([9702621](https://github.com/kurtosis-tech/kurtosis/commit/97026218c036486697bf6b6a8596774a84172b11))

## [0.70.7](https://github.com/kurtosis-tech/kurtosis/compare/0.70.6...0.70.7) (2023-03-28)


### Bug Fixes

* added a fix for release failure ([#361](https://github.com/kurtosis-tech/kurtosis/issues/361)) ([04ee614](https://github.com/kurtosis-tech/kurtosis/commit/04ee6140471a7e6c4b3ea4d6b1e1cb75e4bb1374))

## [0.70.6](https://github.com/kurtosis-tech/kurtosis/compare/0.70.5...0.70.6) (2023-03-28)


### Features

* add search on docs ([#159](https://github.com/kurtosis-tech/kurtosis/issues/159)) ([f80d036](https://github.com/kurtosis-tech/kurtosis/commit/f80d0361435c99707291c0e96c0c308326343330))
* integrate lsp with monrepo ([#223](https://github.com/kurtosis-tech/kurtosis/issues/223)) ([b5ed670](https://github.com/kurtosis-tech/kurtosis/commit/b5ed6707b1c3254cefcfa9201fb76ff1116dedce))


### Bug Fixes

* fix typo in reindex workflow ([#357](https://github.com/kurtosis-tech/kurtosis/issues/357)) ([8900ff2](https://github.com/kurtosis-tech/kurtosis/commit/8900ff230240487e40e706fccc3b8e32b1d1082f))
* remove file paths from error message ([#256](https://github.com/kurtosis-tech/kurtosis/issues/256)) ([cb603d8](https://github.com/kurtosis-tech/kurtosis/commit/cb603d836772812d602bfb86736a8145b85162e1))
* remove trailing errors after starlark execution ([724ac35](https://github.com/kurtosis-tech/kurtosis/commit/724ac355e8cf64a070c3d62a0f593399b5ef2dde))
* remove trailing errors after starlark execution ([#257](https://github.com/kurtosis-tech/kurtosis/issues/257)) ([724ac35](https://github.com/kurtosis-tech/kurtosis/commit/724ac355e8cf64a070c3d62a0f593399b5ef2dde))

## [0.70.5](https://github.com/kurtosis-tech/kurtosis/compare/0.70.4...0.70.5) (2023-03-28)


### Features

* Print the running engine version while running Kurtosis Version ([#346](https://github.com/kurtosis-tech/kurtosis/issues/346)) ([9ef03cb](https://github.com/kurtosis-tech/kurtosis/commit/9ef03cb22e26d9dce556e1d31bacf9b3b30da736))

## [0.70.4](https://github.com/kurtosis-tech/kurtosis/compare/0.70.3...0.70.4) (2023-03-28)


### Features

* Added extra name generation items ([#342](https://github.com/kurtosis-tech/kurtosis/issues/342)) ([0ed2923](https://github.com/kurtosis-tech/kurtosis/commit/0ed2923d9a16cb68b706e25112a741a4b7384944))
* Publish multi-arch image for `files-artifacts-expander` ([#344](https://github.com/kurtosis-tech/kurtosis/issues/344)) ([9e2b369](https://github.com/kurtosis-tech/kurtosis/commit/9e2b369206b974d06e5a7c172b6363e5c6fb1d92))

## [0.70.3](https://github.com/kurtosis-tech/kurtosis/compare/0.70.2...0.70.3) (2023-03-27)


### Features

* Added `bug, feature and docs` flags in the `kurtosis feedback` command ([#287](https://github.com/kurtosis-tech/kurtosis/issues/287)) ([963e9dd](https://github.com/kurtosis-tech/kurtosis/commit/963e9dd055a3ceabafde11a4d942e16f300d827c))


### Bug Fixes

* check service name contains allowed characters and errors cleanly ([#315](https://github.com/kurtosis-tech/kurtosis/issues/315)) ([94af4bd](https://github.com/kurtosis-tech/kurtosis/commit/94af4bda3aac9a1ed45e6ac503407d271ba1d73f)), closes [#164](https://github.com/kurtosis-tech/kurtosis/issues/164)

## [0.70.2](https://github.com/kurtosis-tech/kurtosis/compare/0.70.1...0.70.2) (2023-03-27)


### Features

* Automatically restart engine on context switch ([#329](https://github.com/kurtosis-tech/kurtosis/issues/329)) ([b0712cc](https://github.com/kurtosis-tech/kurtosis/commit/b0712cc42fca1b4ba322bf473da57db4eab8c462))


### Bug Fixes

* Fix info CLI log for portal not running ([#330](https://github.com/kurtosis-tech/kurtosis/issues/330)) ([0fb938e](https://github.com/kurtosis-tech/kurtosis/commit/0fb938e31c29778ac122675e7706e3ad1fd0fc93))

## [0.70.1](https://github.com/kurtosis-tech/kurtosis/compare/0.70.0...0.70.1) (2023-03-27)


### Features

* Add context `rm` command ([#275](https://github.com/kurtosis-tech/kurtosis/issues/275)) ([c20ca12](https://github.com/kurtosis-tech/kurtosis/commit/c20ca121443a78ed6b266cd18b5d69997ee69e57))
* Add context `switch` CLI command ([#317](https://github.com/kurtosis-tech/kurtosis/issues/317)) ([ebab7eb](https://github.com/kurtosis-tech/kurtosis/commit/ebab7ebb697e4c791b47bff14a4e32aaa04268b5))
* add kurtosis engine logs command that dumps logs for all engines in target dir ([#313](https://github.com/kurtosis-tech/kurtosis/issues/313)) ([cbb588c](https://github.com/kurtosis-tech/kurtosis/commit/cbb588c01a6d8d8baffcb41c87b49716c23b7cfc))
* result of add service contains a `name` property ([#314](https://github.com/kurtosis-tech/kurtosis/issues/314)) ([af8ca5f](https://github.com/kurtosis-tech/kurtosis/commit/af8ca5f1d7ec9564baf171ea3478b71e3d9f113b))
* Tunnel remote APIC port to local machine using Kurtosis Portal ([#295](https://github.com/kurtosis-tech/kurtosis/issues/295)) ([4c3ba69](https://github.com/kurtosis-tech/kurtosis/commit/4c3ba69062b78c5f960b4b94fa4427d2b76f6f7a))


### Bug Fixes

* add example historical version ([#150](https://github.com/kurtosis-tech/kurtosis/issues/150)) ([1548489](https://github.com/kurtosis-tech/kurtosis/commit/1548489b87aa927051b4cd01190b92be48e0714d))
* be clear about the engine that is being started ([#282](https://github.com/kurtosis-tech/kurtosis/issues/282)) ([5bc1b79](https://github.com/kurtosis-tech/kurtosis/commit/5bc1b79e94a5688dc908bd413a7f410e8aaf2346))
* Fix starlark value reference bug ([#322](https://github.com/kurtosis-tech/kurtosis/issues/322)) ([63f6626](https://github.com/kurtosis-tech/kurtosis/commit/63f6626be54b71a9fc09e02ae07207c9a9309409))
* name all args for add_services instruction in quickstart ([#316](https://github.com/kurtosis-tech/kurtosis/issues/316)) ([d413826](https://github.com/kurtosis-tech/kurtosis/commit/d41382635d3ad0c51dec6f937c2c26f105fcfe13))
* reformat build prereqs in readme ([#290](https://github.com/kurtosis-tech/kurtosis/issues/290)) ([c286151](https://github.com/kurtosis-tech/kurtosis/commit/c28615144a40cfb369e5fb35d9722ecf912fedce))

## [0.70.0](https://github.com/kurtosis-tech/kurtosis/compare/0.69.2...0.70.0) (2023-03-22)


### ⚠ BREAKING CHANGES

* This is a breaking change where we are removing the ExecRecipe.service_name, GetHttpRequestRecipe.service_name, and PostHttpRequestRecipe.service_name fields, we suggest users pass this value as an argument in the exec, request and wait instructions where this type is currently used. We are also deprecating the previous exec, request, and wait instructions signature that haven't the service_name field, users must add this field on these instructions call. Another change is that now the service_name field on the exec, request, and wait instructions is mandatory ([#301](https://github.com/kurtosis-tech/kurtosis/issues/301))

### Features

* Kurtosis backend can now connect to a remote Docker backend ([#285](https://github.com/kurtosis-tech/kurtosis/issues/285)) ([98b04c8](https://github.com/kurtosis-tech/kurtosis/commit/98b04c8c98e92c0c7e2661ae1020cee1a1cf1e4b))
* This is a breaking change where we are removing the ExecRecipe.service_name, GetHttpRequestRecipe.service_name, and PostHttpRequestRecipe.service_name fields, we suggest users pass this value as an argument in the exec, request and wait instructions where this type is currently used. We are also deprecating the previous exec, request, and wait instructions signature that haven't the service_name field, users must add this field on these instructions call. Another change is that now the service_name field on the exec, request, and wait instructions is mandatory ([#301](https://github.com/kurtosis-tech/kurtosis/issues/301)) ([eb7e88f](https://github.com/kurtosis-tech/kurtosis/commit/eb7e88f3309f6d98e8ddb4ff8aad6baa991ea450))

## [0.69.2](https://github.com/kurtosis-tech/kurtosis/compare/0.69.1...0.69.2) (2023-03-22)


### Features

* Add context `add` command ([#278](https://github.com/kurtosis-tech/kurtosis/issues/278)) ([bd56cae](https://github.com/kurtosis-tech/kurtosis/commit/bd56cae2c7d29dff4c6011ed80521eddfd39277d))
* build script check for go and node versions ([#240](https://github.com/kurtosis-tech/kurtosis/issues/240)) ([4749dbe](https://github.com/kurtosis-tech/kurtosis/commit/4749dbe62030eb17bebd02f81b1a0b822f7e843d))

## [0.69.1](https://github.com/kurtosis-tech/kurtosis/compare/0.69.0...0.69.1) (2023-03-22)


### Features

* Add `context` root command and simple `ls` subcommand ([#241](https://github.com/kurtosis-tech/kurtosis/issues/241)) ([4097c25](https://github.com/kurtosis-tech/kurtosis/commit/4097c25ad57af61f16044b1193df28b5b94acc97))

## [0.69.0](https://github.com/kurtosis-tech/kurtosis/compare/0.68.13...0.69.0) (2023-03-21)


### ⚠ BREAKING CHANGES

* Add acceptable code for request and exec ([#212](https://github.com/kurtosis-tech/kurtosis/issues/212))
* The --enclave-identifier, --enclave-identifiers and --service-identifier flags have been renamed to , --enclave, --enclaves and --service respectively. Users will have to change any scripts or CI configurations that depend on those flags.
* Reduce wait default timeout from 15 minutes to 10 seconds ([#211](https://github.com/kurtosis-tech/kurtosis/issues/211))

### Features

* Add acceptable code for request and exec ([#212](https://github.com/kurtosis-tech/kurtosis/issues/212)) ([9b00ac2](https://github.com/kurtosis-tech/kurtosis/commit/9b00ac2674ce4d602d1eafb4e00e789709917fd5)), closes [#201](https://github.com/kurtosis-tech/kurtosis/issues/201) [#188](https://github.com/kurtosis-tech/kurtosis/issues/188)
* Add library to manage context configurations ([#196](https://github.com/kurtosis-tech/kurtosis/issues/196)) ([c27038a](https://github.com/kurtosis-tech/kurtosis/commit/c27038a41ebb94940881139f990465fffdc0c8d1))
* added a command that allows users to open the Kurtosis Twitter page ([#265](https://github.com/kurtosis-tech/kurtosis/issues/265)) ([c8bcc91](https://github.com/kurtosis-tech/kurtosis/commit/c8bcc91b8f4ff389df216e7f446be10d9100c78c))
* PostHttpRequestRecipe accepts empty body ([#214](https://github.com/kurtosis-tech/kurtosis/issues/214)) ([b7991dc](https://github.com/kurtosis-tech/kurtosis/commit/b7991dc32c31fcac5307d10288bc3908a1b9fc40))
* print files artifacts registered in an enclave during enclave inspect ([#228](https://github.com/kurtosis-tech/kurtosis/issues/228)) ([ef167d6](https://github.com/kurtosis-tech/kurtosis/commit/ef167d692ebac40d60819987d2f11c47fa4658dc))
* Reduce wait default timeout from 15 minutes to 10 seconds ([#211](https://github.com/kurtosis-tech/kurtosis/issues/211)) ([4429284](https://github.com/kurtosis-tech/kurtosis/commit/4429284e35eea6757b22a79a833297ec224c5374))
* rename enclave and service identifier flags ([#264](https://github.com/kurtosis-tech/kurtosis/issues/264)) ([436a44a](https://github.com/kurtosis-tech/kurtosis/commit/436a44a4e4bfa22d9fe5468859f336ecd696c73a))
* update our bug report template ([c84058b](https://github.com/kurtosis-tech/kurtosis/commit/c84058b3e0240893534b150a21cbeb5fb807bfa1))
* update our bug report template ([#237](https://github.com/kurtosis-tech/kurtosis/issues/237)) ([c84058b](https://github.com/kurtosis-tech/kurtosis/commit/c84058b3e0240893534b150a21cbeb5fb807bfa1))


### Bug Fixes

* address typo in our calendly banner link ([#276](https://github.com/kurtosis-tech/kurtosis/issues/276)) ([e1029c3](https://github.com/kurtosis-tech/kurtosis/commit/e1029c3fc41b37468395b16158ef3d0b6cf73082))
* clarify actions for is user-facing changes in PR template ([#279](https://github.com/kurtosis-tech/kurtosis/issues/279)) ([969c3b8](https://github.com/kurtosis-tech/kurtosis/commit/969c3b870bc837b0ee0d6f6e0c1d800cec47419f))
* deprecate --id flag in enclave add ([#247](https://github.com/kurtosis-tech/kurtosis/issues/247)) ([974ff18](https://github.com/kurtosis-tech/kurtosis/commit/974ff186478499806156a08772ec9bc997665b31))
* Lock default context in contexts config ([#277](https://github.com/kurtosis-tech/kurtosis/issues/277)) ([8da3b94](https://github.com/kurtosis-tech/kurtosis/commit/8da3b94405e6d5e5f1fe659b137287e97ceb061d))
* Update PR title workflow for merge queue ([#267](https://github.com/kurtosis-tech/kurtosis/issues/267)) ([00ccfec](https://github.com/kurtosis-tech/kurtosis/commit/00ccfecf5d26ee440010c4a6ffd32f7dd7b15d8b))
* warn if engine version is greater than cli and error if cli &gt; engine ([#243](https://github.com/kurtosis-tech/kurtosis/issues/243)) ([03352e1](https://github.com/kurtosis-tech/kurtosis/commit/03352e128c6521b32e48f4036cbfe4ba803fbf84))

## [0.68.13](https://github.com/kurtosis-tech/kurtosis/compare/0.68.12...0.68.13) (2023-03-16)


### Features

* made the content-type field optional in PostHttpRequestRecipe ([#222](https://github.com/kurtosis-tech/kurtosis/issues/222)) ([d551398](https://github.com/kurtosis-tech/kurtosis/commit/d551398112aded68dd348c661fb14512080a9bdb))


### Bug Fixes

* add trailing commas to Starlark code ([#218](https://github.com/kurtosis-tech/kurtosis/issues/218)) ([1bd050c](https://github.com/kurtosis-tech/kurtosis/commit/1bd050c8de01fd24bae5ffaf786aa87b86bdf134))
* collapse current behavior into background+motivation ([#216](https://github.com/kurtosis-tech/kurtosis/issues/216)) ([853aa5d](https://github.com/kurtosis-tech/kurtosis/commit/853aa5d9ee79b7f540897f2ca0ac80f5c31740ec))
* print the upgrade CLI warning at most hourly ([#224](https://github.com/kurtosis-tech/kurtosis/issues/224)) ([f40ee90](https://github.com/kurtosis-tech/kurtosis/commit/f40ee90c4d1008a932daa902a264acf3e4b48510))
* refer to the new repo name in remote subpackage tests ([#225](https://github.com/kurtosis-tech/kurtosis/issues/225)) ([cd81f2e](https://github.com/kurtosis-tech/kurtosis/commit/cd81f2ef8d721e94dd0b0c668d9ddaf64b03677d))

## [0.68.12](https://github.com/kurtosis-tech/kurtosis/compare/0.68.11...0.68.12) (2023-03-15)


### Bug Fixes

* wait instruction hanging forever when `service_name` field is not passed ([#197](https://github.com/kurtosis-tech/kurtosis/issues/197)) ([826f072](https://github.com/kurtosis-tech/kurtosis/commit/826f0727a43ca1acc05aaded41eed307b04c7d96))

## [0.68.11](https://github.com/kurtosis-tech/kurtosis/compare/0.68.10...0.68.11) (2023-03-15)


### Features

* colorize RUNNING|STOPPED statuses for Enclaves And Containers ([#178](https://github.com/kurtosis-tech/kurtosis/issues/178)) ([8254c7f](https://github.com/kurtosis-tech/kurtosis/commit/8254c7fbf35e38840c1ff5182017f19184eccae5))


### Bug Fixes

* remove api container stuff & colorize keys ([#195](https://github.com/kurtosis-tech/kurtosis/issues/195)) ([9ccb910](https://github.com/kurtosis-tech/kurtosis/commit/9ccb9102736eda2e8cb6645796cb9bfc73209ea1))

## [0.68.10](https://github.com/kurtosis-tech/kurtosis/compare/0.68.9...0.68.10) (2023-03-15)


### Bug Fixes

* Tag docker images correctly after Kudet removal ([#206](https://github.com/kurtosis-tech/kurtosis/issues/206)) ([2e594a4](https://github.com/kurtosis-tech/kurtosis/commit/2e594a444a2eef5b058402edf675b7526a0ec675))

## [0.68.9](https://github.com/kurtosis-tech/kurtosis/compare/0.68.8...0.68.9) (2023-03-15)


### Features

* Add a new pull request template ([#117](https://github.com/kurtosis-tech/kurtosis/issues/117)) ([45b2067](https://github.com/kurtosis-tech/kurtosis/commit/45b2067302f9fb38c2dda43dedbdbbcc7878fea6))
* show enclave inspect immediately after run ([#170](https://github.com/kurtosis-tech/kurtosis/issues/170)) ([5790131](https://github.com/kurtosis-tech/kurtosis/commit/57901311eefdbe877e97deef4ee3e5ba1bd4c75a))


### Bug Fixes

* Add back fetch depth to change version GH action ([f5f32a2](https://github.com/kurtosis-tech/kurtosis/commit/f5f32a294fdf365cde2e998b03e37ab1a1b42d14))
* Add back fetch depth to change version GH action ([#204](https://github.com/kurtosis-tech/kurtosis/issues/204)) ([f5f32a2](https://github.com/kurtosis-tech/kurtosis/commit/f5f32a294fdf365cde2e998b03e37ab1a1b42d14))
* remove & service uuid from autocomplete ([#182](https://github.com/kurtosis-tech/kurtosis/issues/182)) ([3be2070](https://github.com/kurtosis-tech/kurtosis/commit/3be207091fcb99161a7e8b8712d885a3c1298954))
* use with-subnetworks ([#163](https://github.com/kurtosis-tech/kurtosis/issues/163)) ([db6dd41](https://github.com/kurtosis-tech/kurtosis/commit/db6dd41e7415d30d0811516525395010bb02c6d5))

## [0.68.8](https://github.com/kurtosis-tech/kurtosis/compare/0.68.7...0.68.8) (2023-03-14)


### Bug Fixes

* bump historical cli install down the sidebar ([cba11eb](https://github.com/kurtosis-tech/kurtosis/commit/cba11eb3fe5545166b4979aeb63e2c26dd3c375b))
* bump historical cli install down the sidebar ([#152](https://github.com/kurtosis-tech/kurtosis/issues/152)) ([cba11eb](https://github.com/kurtosis-tech/kurtosis/commit/cba11eb3fe5545166b4979aeb63e2c26dd3c375b))
* print enclave names even after restart during clean ([#156](https://github.com/kurtosis-tech/kurtosis/issues/156)) ([43ab71e](https://github.com/kurtosis-tech/kurtosis/commit/43ab71e3305f3c434f6d5718e4e2d2b664993ae2))

## [0.68.7](https://github.com/kurtosis-tech/kurtosis/compare/0.68.6...0.68.7) (2023-03-13)


### Bug Fixes

* added instruction position while executing starlark package ([bc70e4e](https://github.com/kurtosis-tech/kurtosis/commit/bc70e4e1b5ad743edf9dcaa7b0feb0975e8f7eb0))
* added instruction position while executing starlark package ([#143](https://github.com/kurtosis-tech/kurtosis/issues/143)) ([bc70e4e](https://github.com/kurtosis-tech/kurtosis/commit/bc70e4e1b5ad743edf9dcaa7b0feb0975e8f7eb0))
* fix changelog for versioned docs going forward ([#142](https://github.com/kurtosis-tech/kurtosis/issues/142)) ([2fc3e72](https://github.com/kurtosis-tech/kurtosis/commit/2fc3e72248bbbbb1780ecf32db95a6c9fbe08972))
* gramatical fix in analytics tracking logging ([#138](https://github.com/kurtosis-tech/kurtosis/issues/138)) ([23212a3](https://github.com/kurtosis-tech/kurtosis/commit/23212a3188445e3f358eef0e3ac388752eb9a0c7))
* sort services by name ([#139](https://github.com/kurtosis-tech/kurtosis/issues/139)) ([d60ef67](https://github.com/kurtosis-tech/kurtosis/commit/d60ef67e0fa2e456d11b0a3925dd731a969928d6))

## [0.68.6](https://github.com/kurtosis-tech/kurtosis/compare/0.68.5...0.68.6) (2023-03-09)


### Features

* Added `kurtosis feedback` CLI command ([#28](https://github.com/kurtosis-tech/kurtosis/issues/28)) ([55210ec](https://github.com/kurtosis-tech/kurtosis/commit/55210ec5660f6c642eda4baa422cf766fc584be5))
* publish versioned brew formula ([#130](https://github.com/kurtosis-tech/kurtosis/issues/130)) ([a7d695d](https://github.com/kurtosis-tech/kurtosis/commit/a7d695d3fc58d7c4c3c3fd218bf9af98a3bc0086))

## [0.68.5](https://github.com/kurtosis-tech/kurtosis/compare/0.68.4...0.68.5) (2023-03-09)


### Bug Fixes

* Use version.txt for kurtosis_version instead of Git tags ([#126](https://github.com/kurtosis-tech/kurtosis/issues/126)) ([f5bfe9e](https://github.com/kurtosis-tech/kurtosis/commit/f5bfe9e5795305c172a6fd02115825b2ea0b638a))

## [0.68.4](https://github.com/kurtosis-tech/kurtosis/compare/0.68.3...0.68.4) (2023-03-09)


### Bug Fixes

* Pass correct latest tag to GoReleaser CLI build ([#122](https://github.com/kurtosis-tech/kurtosis/issues/122)) ([ec10c54](https://github.com/kurtosis-tech/kurtosis/commit/ec10c542d2ef97dd4c3ca0d542fa5af23fc44ca2))

## [0.68.3](https://github.com/kurtosis-tech/kurtosis/compare/0.68.2...0.68.3) (2023-03-08)


### Features

* Use semver versioning for Golang API package ([#119](https://github.com/kurtosis-tech/kurtosis/issues/119)) ([1d4ff7f](https://github.com/kurtosis-tech/kurtosis/commit/1d4ff7fea55bcf25538b955275d776ff0b2f3678))


### Bug Fixes

* remove mentions about github discussions ([#95](https://github.com/kurtosis-tech/kurtosis/issues/95)) ([2387fa2](https://github.com/kurtosis-tech/kurtosis/commit/2387fa230bc5a6d240755acbbb9b5cbcc5489ea0))

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

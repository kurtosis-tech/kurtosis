# Changelog

## [0.66.6](https://github.com/kurtosis-tech/kurtosis/compare/0.66.5...0.66.6) (2023-02-14)


### Bug Fixes

* Fix broken release of CLI ([#1066](https://github.com/kurtosis-tech/kurtosis/issues/1066)) ([68fc458](https://github.com/kurtosis-tech/kurtosis/commit/68fc458d984929bf5a489942e111bf0970ab2ac0))

## [0.66.5](https://github.com/kurtosis-tech/kurtosis/compare/0.66.4...0.66.5) (2023-02-14)


### Features

* Make Linux CLI binary support plugins ([#1047](https://github.com/kurtosis-tech/kurtosis/issues/1047)) ([b24d68e](https://github.com/kurtosis-tech/kurtosis/commit/b24d68e1409981c2db0e0ecbd21b64eaecb50dc9))

## [0.66.4](https://github.com/kurtosis-tech/kurtosis/compare/0.66.3...0.66.4) (2023-02-11)


### Features

* allow starlark to run subpackages in a repository  ([#995](https://github.com/kurtosis-tech/kurtosis/issues/995)) ([f4ab6e4](https://github.com/kurtosis-tech/kurtosis/commit/f4ab6e4f2885daca4a20b557d52ae46b292842f4))
* kurtosis package in subfolder can be run locally or remotely ([f4ab6e4](https://github.com/kurtosis-tech/kurtosis/commit/f4ab6e4f2885daca4a20b557d52ae46b292842f4))

## [0.66.3](https://github.com/kurtosis-tech/kurtosis/compare/0.66.2...0.66.3) (2023-02-09)


### Features

* add the ability to run non master branch packages ([#1019](https://github.com/kurtosis-tech/kurtosis/issues/1019)) ([245c63a](https://github.com/kurtosis-tech/kurtosis/commit/245c63a9ebba21f5ee6f4b17cc500932c3e5d911)), closes [#422](https://github.com/kurtosis-tech/kurtosis/issues/422)

## [0.66.2](https://github.com/kurtosis-tech/kurtosis/compare/0.66.1...0.66.2) (2023-02-09)


### Bug Fixes

* Add musl-dev dependency to publish tasks ([#1031](https://github.com/kurtosis-tech/kurtosis/issues/1031)) ([ce50380](https://github.com/kurtosis-tech/kurtosis/commit/ce503807fbb51c6d1bbb368f8868e4fca0d17d97))

## [0.66.1](https://github.com/kurtosis-tech/kurtosis/compare/0.66.0...0.66.1) (2023-02-09)


### Bug Fixes

* explode cli docs for improve readability ([#1015](https://github.com/kurtosis-tech/kurtosis/issues/1015)) ([4c60e45](https://github.com/kurtosis-tech/kurtosis/commit/4c60e4573cc0cf744e38f295156c7a7b0411a36e))
* Keep the file-artifact-expander container if something goes wrong ([#1026](https://github.com/kurtosis-tech/kurtosis/issues/1026)) ([be394dc](https://github.com/kurtosis-tech/kurtosis/commit/be394dc9d2dc624a8a85ff6b94e3c1666f424383))

## [0.66.0](https://github.com/kurtosis-tech/kurtosis/compare/0.65.1...0.66.0) (2023-02-07)


### ⚠ BREAKING CHANGES

* We now create a logs collector per enclave instead of per engine. Any enclaves using the old logs collector will not have any new logs published. Any historical enclaves using the old logs collector cannot have services added either.

### Features

* Added an 'Examples' link to [the examples repo](https://github.com/kurtosis-tech/examples) in the sidebar ([6dbfb89](https://github.com/kurtosis-tech/kurtosis/commit/6dbfb89aecd4e35b799ca161be406432f1d6baf5))
* Configurable parallelism for Starlark ([#996](https://github.com/kurtosis-tech/kurtosis/issues/996)) ([46e2c74](https://github.com/kurtosis-tech/kurtosis/commit/46e2c74b903fa144de96ea76f390312656cb889f))
* packet delays can be configured as a distribution between subnetworks using UniformPacketDelayDistribution and NormalPacketDelayDistribution ([#988](https://github.com/kurtosis-tech/kurtosis/issues/988)) ([73dacf4](https://github.com/kurtosis-tech/kurtosis/commit/73dacf463fbbc9f584e6987ba1d3e8a92bc1f5e4))


### Bug Fixes

* remove non determinism around bridge network for user-services by removing access to bridge network ([#991](https://github.com/kurtosis-tech/kurtosis/issues/991)) ([43d4710](https://github.com/kurtosis-tech/kurtosis/commit/43d47109c4c444a5e429344e309df7c5c7300e75))
* removes the 'edf' prefix from enclave uuids ([43d4710](https://github.com/kurtosis-tech/kurtosis/commit/43d47109c4c444a5e429344e309df7c5c7300e75))


### Code Refactoring

* create logs collector per enclave ([#989](https://github.com/kurtosis-tech/kurtosis/issues/989)) ([bc5e894](https://github.com/kurtosis-tech/kurtosis/commit/bc5e894428d2b58c2e695ca4c7c7c8fa1176867e))

## [0.65.1](https://github.com/kurtosis-tech/kurtosis/compare/0.65.0...0.65.1) (2023-02-01)


### Features

* Add `hostname` attribute to the `service` object returned by the `add_service` instruction ([#844](https://github.com/kurtosis-tech/kurtosis/issues/844)) ([50e786c](https://github.com/kurtosis-tech/kurtosis/commit/50e786c29ee6b424dd20988a13f97d1896260c88)), closes [#745](https://github.com/kurtosis-tech/kurtosis/issues/745)
* add typescript SDK for historical service & enclave identifiers ([#923](https://github.com/kurtosis-tech/kurtosis/issues/923)) ([0fc8705](https://github.com/kurtosis-tech/kurtosis/commit/0fc870546a8381f0346b880d93555c3be352f946))
* print shortened uuid by default for add service ([#957](https://github.com/kurtosis-tech/kurtosis/issues/957)) ([057ba5c](https://github.com/kurtosis-tech/kurtosis/commit/057ba5c725fcc62ae5173cf85cc8d1d4fa5fe68d))


### Bug Fixes

* (docs) update links to "enclave" and "services" in the Resource Identifier page in the docs ([#977](https://github.com/kurtosis-tech/kurtosis/issues/977)) ([a7cc76d](https://github.com/kurtosis-tech/kurtosis/commit/a7cc76de0a64f4a29b500009328dde0dc8fa2813))
* add M1 support to Bash completion instructions ([#961](https://github.com/kurtosis-tech/kurtosis/issues/961)) ([de41684](https://github.com/kurtosis-tech/kurtosis/commit/de416848a33c44789d42d71a5d3d8e54d1a5e81d))
* support shorter uuids from previous versions of Kurtosis ([#985](https://github.com/kurtosis-tech/kurtosis/issues/985)) ([3f8b6fc](https://github.com/kurtosis-tech/kurtosis/commit/3f8b6fc712d6b9fdf899f8d03c741d5efdf9fbc6))

## [0.65.0](https://github.com/kurtosis-tech/kurtosis/compare/0.64.2...0.65.0) (2023-01-26)


### ⚠ BREAKING CHANGES

* Remove the backward compatibility support for recipes in favour of pre-defined kurtosis types. ([#920](https://github.com/kurtosis-tech/kurtosis/issues/920))
* This removes backwards compatibility for artifact_id in favor of artifact_name
* Remove support for unnamed `struct` as `add_service` `config` argument. All `add_service` instructions should now use a proper `ServiceConfig` object as their second argument. ([#778](https://github.com/kurtosis-tech/kurtosis/issues/778))
* service_name in Starlark has been renamed service_id to for instructions add_service, remove_service, store_service_files and update_service. This also applies to the exec, post and get recipes. Users of these instructions and recipes will have to service_id rename to service_name for each named usage of service_id to service_name. If you only have positional usage then & go unaffected.

### Features

* `add_services` now fails as soon as one service from the batch fails ([#934](https://github.com/kurtosis-tech/kurtosis/issues/934)) ([ae2fc27](https://github.com/kurtosis-tech/kurtosis/commit/ae2fc2720c7e9368523e8599ed933ff442ee2763)), closes [#933](https://github.com/kurtosis-tech/kurtosis/issues/933)
* Add `add_services` instruction to start services in bulk ([#912](https://github.com/kurtosis-tech/kurtosis/issues/912)) ([e3c3124](https://github.com/kurtosis-tech/kurtosis/commit/e3c3124e54ec5448b13842cd276b349e349bcb0e)), closes [#802](https://github.com/kurtosis-tech/kurtosis/issues/802)
* added a `--full-uuid` flag that prints uuids fully otherwise prints short uuids ([#898](https://github.com/kurtosis-tech/kurtosis/issues/898)) ([9b10342](https://github.com/kurtosis-tech/kurtosis/commit/9b103429bd0a61dc57d102fac618dfdbf11f07d3)), closes [#896](https://github.com/kurtosis-tech/kurtosis/issues/896)
* added default issue template ([#869](https://github.com/kurtosis-tech/kurtosis/issues/869)) ([8f141e3](https://github.com/kurtosis-tech/kurtosis/commit/8f141e3489f949f55f3e73539cf511eec7f0cdc7))
* added packet latency functionality ([#888](https://github.com/kurtosis-tech/kurtosis/issues/888)) ([6d28b54](https://github.com/kurtosis-tech/kurtosis/commit/6d28b5456c5a80e7804ac4ec23d1b04638f8147b))
* added the ability for users to introduce delay between subnetworks ([#897](https://github.com/kurtosis-tech/kurtosis/issues/897)) ([3c5c841](https://github.com/kurtosis-tech/kurtosis/commit/3c5c8418c2e7eee6d35de736970143585b142f0a))
* Remove the backward compatibility support for recipes in favour of pre-defined kurtosis types. ([#920](https://github.com/kurtosis-tech/kurtosis/issues/920)) ([d29455c](https://github.com/kurtosis-tech/kurtosis/commit/d29455c1eaa86ceec39f74f4c096ab7fccd485ca))
* support historical identifiers for enclaves & services for logs ([#900](https://github.com/kurtosis-tech/kurtosis/issues/900)) ([4db5f1e](https://github.com/kurtosis-tech/kurtosis/commit/4db5f1edbdc0999bd59a89397b7a7dc0ca0c30bc))
* Support uuids, names and shortened uuids for enclaves ([#827](https://github.com/kurtosis-tech/kurtosis/issues/827)) ([60f32bf](https://github.com/kurtosis-tech/kurtosis/commit/60f32bf60d8f7205b2b1b4c57a8984a40dcfb664)), closes [#310](https://github.com/kurtosis-tech/kurtosis/issues/310)


### Bug Fixes

* align convertApiPortsToServiceContextPorts functionality across golang and typescript ([#819](https://github.com/kurtosis-tech/kurtosis/issues/819)) ([e7f3425](https://github.com/kurtosis-tech/kurtosis/commit/e7f34259277861cde4653085565455e3b7a35f4d)), closes [#18](https://github.com/kurtosis-tech/kurtosis/issues/18)
* Fix instructions syntax in the quickstart documentation ([#892](https://github.com/kurtosis-tech/kurtosis/issues/892)) ([70d93fd](https://github.com/kurtosis-tech/kurtosis/commit/70d93fd45c05feaef2a480ea540b48912984565c))
* Fix magic string replacement in ServiceConfig ([#942](https://github.com/kurtosis-tech/kurtosis/issues/942)) ([8aeb8fe](https://github.com/kurtosis-tech/kurtosis/commit/8aeb8fed432b300f3d49ea8e29379ce63416710e))
* minor corrections in `add_services` docs ([#924](https://github.com/kurtosis-tech/kurtosis/issues/924)) ([0c780ce](https://github.com/kurtosis-tech/kurtosis/commit/0c780ce53d401fa0d0a59aaf0eb63ea1ba5034dc))
* Only publish via CI on master ([#905](https://github.com/kurtosis-tech/kurtosis/issues/905)) ([42fe3be](https://github.com/kurtosis-tech/kurtosis/commit/42fe3be946783a99f444f710951f11cbca4f9d30))
* rename --full-uuid to --full-uuids ([#941](https://github.com/kurtosis-tech/kurtosis/issues/941)) ([7578296](https://github.com/kurtosis-tech/kurtosis/commit/7578296040c6fd038d501267ebea0cde0f2a3ada))
* rename SDK call and correct docs around fetching of historical identifiers ([#932](https://github.com/kurtosis-tech/kurtosis/issues/932)) ([ff8d515](https://github.com/kurtosis-tech/kurtosis/commit/ff8d5155ac0147a99b977576146f3711897284d3))
* return enclave name instead of enclave uuid for enclave_context in typescript ([#921](https://github.com/kurtosis-tech/kurtosis/issues/921)) ([537475e](https://github.com/kurtosis-tech/kurtosis/commit/537475e1f8ef72bbe39fbe4e763b9c42221630ad))


### Code Refactoring

* remove backwards compatibility for artifact_id ([#915](https://github.com/kurtosis-tech/kurtosis/issues/915)) ([7fcfefc](https://github.com/kurtosis-tech/kurtosis/commit/7fcfefcf6a2ab9f0b7f7040a682999240f07e48a)), closes [#829](https://github.com/kurtosis-tech/kurtosis/issues/829)
* Remove support for unnamed `struct` as `add_service` `config` argument. All `add_service` instructions should now use a proper `ServiceConfig` object as their second argument. ([#778](https://github.com/kurtosis-tech/kurtosis/issues/778)) ([b402e36](https://github.com/kurtosis-tech/kurtosis/commit/b402e36f6582db55f0d68875db221aaeddcaab19))
* rename `service_id` in Starlark to `service_name` ([#890](https://github.com/kurtosis-tech/kurtosis/issues/890)) ([a27d9d4](https://github.com/kurtosis-tech/kurtosis/commit/a27d9d4c69b1b519770cadc8884b16d65477164d)), closes [#861](https://github.com/kurtosis-tech/kurtosis/issues/861)

## [0.64.2](https://github.com/kurtosis-tech/kurtosis/compare/0.64.1...0.64.2) (2023-01-17)


### Features

* added backward compatibility to the the new constructors ([2c319b3](https://github.com/kurtosis-tech/kurtosis/commit/2c319b30ab92b126c2f815f6b6448bd464f3f167))
* added http and exec recipe constructor ([2c319b3](https://github.com/kurtosis-tech/kurtosis/commit/2c319b30ab92b126c2f815f6b6448bd464f3f167))
* added http recipe constructor ([#794](https://github.com/kurtosis-tech/kurtosis/issues/794)) ([2c319b3](https://github.com/kurtosis-tech/kurtosis/commit/2c319b30ab92b126c2f815f6b6448bd464f3f167))


### Bug Fixes

* added cpu in starlark-instruction ([0bd0dfe](https://github.com/kurtosis-tech/kurtosis/commit/0bd0dfe881cc0b2935ae6457abb179f038b59a64))
* created separate constructors for Post and Get request recipes. ([2c319b3](https://github.com/kurtosis-tech/kurtosis/commit/2c319b30ab92b126c2f815f6b6448bd464f3f167))
* update add-service link ([0bd0dfe](https://github.com/kurtosis-tech/kurtosis/commit/0bd0dfe881cc0b2935ae6457abb179f038b59a64))

## [0.64.1](https://github.com/kurtosis-tech/kurtosis/compare/0.64.0...0.64.1) (2023-01-11)


### Features

* allow downloading artifacts ([#840](https://github.com/kurtosis-tech/kurtosis/issues/840)) ([046e0b0](https://github.com/kurtosis-tech/kurtosis/commit/046e0b0ef1485c792b0624b91b67b84602886494)), closes [#832](https://github.com/kurtosis-tech/kurtosis/issues/832)


### Bug Fixes

* Fix the serialization of the `Service` object returned by the `add_service` instruction ([#845](https://github.com/kurtosis-tech/kurtosis/issues/845)) ([7fe5ec0](https://github.com/kurtosis-tech/kurtosis/commit/7fe5ec0ceed0845e705215400a8c3be5b90512b6)), closes [#721](https://github.com/kurtosis-tech/kurtosis/issues/721)
* Make PR description update support dependabot PRs ([8e1c2da](https://github.com/kurtosis-tech/kurtosis/commit/8e1c2da5b2a2222b1bd52f21ae975afbf384ee24))
* Make PR description update support dependabot PRs ([#842](https://github.com/kurtosis-tech/kurtosis/issues/842)) ([8e1c2da](https://github.com/kurtosis-tech/kurtosis/commit/8e1c2da5b2a2222b1bd52f21ae975afbf384ee24))
* The multiline output of an recipe passed to an exec command is not printed on multiple lines ([#841](https://github.com/kurtosis-tech/kurtosis/issues/841)) ([700eb9d](https://github.com/kurtosis-tech/kurtosis/commit/700eb9da4534db4f3bdbc323005977e422cb4ef2)), closes [#732](https://github.com/kurtosis-tech/kurtosis/issues/732)

## [0.64.0](https://github.com/kurtosis-tech/kurtosis/compare/0.63.2...0.64.0) (2023-01-10)


### ⚠ BREAKING CHANGES

* Files artifacts now support both names and uuids. Names provided using (artifact_id, name) are mandatory in Starlark. Note that soon we will be dropping support for artifact_id in favor of name. This applies to upload_files, render_templates and store_service_files. In the Go or TS SDK artifactName has also been introduced and made mandatory in the uploadFiles and storeWebFiles functions. Users must choose a point in time unique name for their artifacts.
* Deprecate AddService SDK call ([#803](https://github.com/kurtosis-tech/kurtosis/issues/803))

### Features

* **!:** Update `wait` backoff to be a constant backoff. The `interval` argument of `wait` is now the constant interval, not the initial interval for an "Exponential Backoff" strategy. You might want to increase it slightly if it is set to a small duration value ([#793](https://github.com/kurtosis-tech/kurtosis/issues/793)) ([4f728c0](https://github.com/kurtosis-tech/kurtosis/commit/4f728c0ec44d1202a94783b4bf8f3f25ee818670))
* Add CPU and Memory allocation to ([27f505e](https://github.com/kurtosis-tech/kurtosis/commit/27f505e0ff3327a170236013d27357c982c0adcf))
* Add CPU and Memory allocation to add_service ([#790](https://github.com/kurtosis-tech/kurtosis/issues/790)) ([27f505e](https://github.com/kurtosis-tech/kurtosis/commit/27f505e0ff3327a170236013d27357c982c0adcf))
* Assert can take two runtime values ([#787](https://github.com/kurtosis-tech/kurtosis/issues/787)) ([c0bb124](https://github.com/kurtosis-tech/kurtosis/commit/c0bb124f911484f65fd658a8433a547e1a34a385))
* support passing names for cli generated artifacts ([#834](https://github.com/kurtosis-tech/kurtosis/issues/834)) ([0cc8fb3](https://github.com/kurtosis-tech/kurtosis/commit/0cc8fb3ea0e5c2a071a682941b53aba72116c9d1))


### Bug Fixes

* disabled for typescripts tests as well ([97a0c1e](https://github.com/kurtosis-tech/kurtosis/commit/97a0c1e3998f309d9f590c975b6a4e3a3421ba3b))
* Fix starlark command for CLI rendertemplate ([f8542dd](https://github.com/kurtosis-tech/kurtosis/commit/f8542dd5626494ba9f19a2df3872c7a0b509700e))
* Fix starlark command for CLI storeservice ([f8542dd](https://github.com/kurtosis-tech/kurtosis/commit/f8542dd5626494ba9f19a2df3872c7a0b509700e))
* Fix starlark scripts for CLI commands ([#824](https://github.com/kurtosis-tech/kurtosis/issues/824)) ([f8542dd](https://github.com/kurtosis-tech/kurtosis/commit/f8542dd5626494ba9f19a2df3872c7a0b509700e))
* fixed search-logs-test in the internal testsuite ([#811](https://github.com/kurtosis-tech/kurtosis/issues/811)) ([3196e6c](https://github.com/kurtosis-tech/kurtosis/commit/3196e6cbbe4a574cdf92bff93d10c28b0bbd6b35))
* removed the dependency on file-server static in tests ([97a0c1e](https://github.com/kurtosis-tech/kurtosis/commit/97a0c1e3998f309d9f590c975b6a4e3a3421ba3b))
* removed the dependency on file-server static in tests ([#814](https://github.com/kurtosis-tech/kurtosis/issues/814)) ([97a0c1e](https://github.com/kurtosis-tech/kurtosis/commit/97a0c1e3998f309d9f590c975b6a4e3a3421ba3b))
* Wait returns a nice message when it times out ([#788](https://github.com/kurtosis-tech/kurtosis/issues/788)) ([9e402ba](https://github.com/kurtosis-tech/kurtosis/commit/9e402ba5ab1f8594302cd18268611f04187566dc))


### Code Refactoring

* Deprecate AddService SDK call ([#803](https://github.com/kurtosis-tech/kurtosis/issues/803)) ([89bf3e7](https://github.com/kurtosis-tech/kurtosis/commit/89bf3e7d37184b96cb4ecff9df759fa7f23c13b7)), closes [#460](https://github.com/kurtosis-tech/kurtosis/issues/460) [#443](https://github.com/kurtosis-tech/kurtosis/issues/443)
* Files artifacts support names and uuids ([#804](https://github.com/kurtosis-tech/kurtosis/issues/804)) ([d35299d](https://github.com/kurtosis-tech/kurtosis/commit/d35299dca0e5d34974277f271cb773ae3898a5ff)), closes [#799](https://github.com/kurtosis-tech/kurtosis/issues/799)

## [0.63.2](https://github.com/kurtosis-tech/kurtosis/compare/0.63.1...0.63.2) (2022-12-21)


### Bug Fixes

* Fix `ServiceConfig` attributes for `hasattr` to be backward compatible ([#779](https://github.com/kurtosis-tech/kurtosis/issues/779)) ([1361292](https://github.com/kurtosis-tech/kurtosis/commit/136129220e9d13f084b33db4ce174dcd8dfe319e))

## [0.63.1](https://github.com/kurtosis-tech/kurtosis/compare/0.63.0...0.63.1) (2022-12-20)


### Features

* Add a `ServiceConfig` for the add_service instruction. All `add_service` instructions should update their `config` argument to use a ServiceConfig object rather than an unnamed `struct`. Support for `struct` will be removed shortly. ([#757](https://github.com/kurtosis-tech/kurtosis/issues/757)) ([de48639](https://github.com/kurtosis-tech/kurtosis/commit/de486394978addd8230f4aa937b93b42bac9361e))


### Bug Fixes

* fixed and issue when the start time request values was not properly set when finding for existing service GUIDs before streaming logs ([#755](https://github.com/kurtosis-tech/kurtosis/issues/755)) ([995381a](https://github.com/kurtosis-tech/kurtosis/commit/995381a71b210a67cde3c2297c7160be4de24559))

## [0.63.0](https://github.com/kurtosis-tech/kurtosis/compare/0.62.0...0.63.0) (2022-12-20)

### ⚠ BREAKING CHANGES

* Make a plan argument required for run. Users will have to now pass a `plan` object to the `run` function in the main.star as the first argument. If you are passing args the argument should be called `args` and has to be the second argument. Users will have to futher use all enclave-modifying functions like add_service, remove_service from the `plan` object. ([#728](https://github.com/kurtosis-tech/kurtosis/issues/728)) ([a33c0aa](https://github.com/kurtosis-tech/kurtosis/commit/a33c0aaca4ded827e6f92a9c200689540f212146))

### Features

* make a plan argument required for run ([#728](https://github.com/kurtosis-tech/kurtosis/issues/728)) ([a33c0aa](https://github.com/kurtosis-tech/kurtosis/commit/a33c0aaca4ded827e6f92a9c200689540f212146))

## [0.62.0](https://github.com/kurtosis-tech/kurtosis/compare/0.61.0...0.62.0) (2022-12-16)


### ⚠ BREAKING CHANGES

* Kurtosis subnetwork capabilities are available in Starlark ([#734](https://github.com/kurtosis-tech/kurtosis/issues/734))
* invert the order of files and artifact ids passed to files artifacts. users will have to invert the dictionary in their starlark code. ([#711](https://github.com/kurtosis-tech/kurtosis/issues/711))

### Features

* `add_service` now accepts `subnetwork` inside its config argument ([#670](https://github.com/kurtosis-tech/kurtosis/issues/670)) ([996c24b](https://github.com/kurtosis-tech/kurtosis/commit/996c24b5941db06cf765735475bba73e82297fb0))
* Add `kurtosis` module containing `connection.[BLOCKED|ALLOWED]` ([#718](https://github.com/kurtosis-tech/kurtosis/issues/718)) ([b71ee95](https://github.com/kurtosis-tech/kurtosis/commit/b71ee9598bfa857200351d2f270842366374f51a))
* Add recipe support to exec command  ([#668](https://github.com/kurtosis-tech/kurtosis/issues/668)) ([c8fd7c1](https://github.com/kurtosis-tech/kurtosis/commit/c8fd7c1c3e1106ce4c946a401c15d12132651cf8)), closes [#510](https://github.com/kurtosis-tech/kurtosis/issues/510) [#627](https://github.com/kurtosis-tech/kurtosis/issues/627)
* Add support for exec recipe on wait command ([#700](https://github.com/kurtosis-tech/kurtosis/issues/700)) ([b9bc1d0](https://github.com/kurtosis-tech/kurtosis/commit/b9bc1d087a6e7c51a474ecd336c27e603c7a830c)), closes [#698](https://github.com/kurtosis-tech/kurtosis/issues/698)
* Added `match`, `regex-match` and `invert-match` flags in the `search logs` CLI command to allow users to filter the returned log lines ([#717](https://github.com/kurtosis-tech/kurtosis/issues/717)) ([4a3e814](https://github.com/kurtosis-tech/kurtosis/commit/4a3e814e1ddf8b09d72e2e757b7269f0674e95ce))
* Added `remove_connection` instruction to remove a connection between 2 subnetworks ([#692](https://github.com/kurtosis-tech/kurtosis/issues/692)) ([5905cc3](https://github.com/kurtosis-tech/kurtosis/commit/5905cc3f26a305557d25b9093b45c69bd4d3f288))
* Added `update-service` instruction to move a service from a subnetwork to a different subnetwork ([#715](https://github.com/kurtosis-tech/kurtosis/issues/715)) ([d652a5d](https://github.com/kurtosis-tech/kurtosis/commit/d652a5dd70976f3b95650a82f43ca263f491386c))
* Added validation that will fail early if subnetwork feature is being used in an enclave that does not support it ([#731](https://github.com/kurtosis-tech/kurtosis/issues/731)) ([9c3e869](https://github.com/kurtosis-tech/kurtosis/commit/9c3e869e03776ac14058e47c5f7ed16b9f364419))
* invert the order of files and artifact ids passed to files artifacts. users will have to invert the dictionary in their starlark code. ([#711](https://github.com/kurtosis-tech/kurtosis/issues/711)) ([9acfe3e](https://github.com/kurtosis-tech/kurtosis/commit/9acfe3e08bbb310da903980891c520236803711b)), closes [#545](https://github.com/kurtosis-tech/kurtosis/issues/545)
* Kurtosis subnetwork capabilities are available in Starlark ([#734](https://github.com/kurtosis-tech/kurtosis/issues/734)) ([cfacf16](https://github.com/kurtosis-tech/kurtosis/commit/cfacf16447f4da3d96b53baed3550c34c7fe12bd))
* support runtime values and ip addresses in exec and request ([f09b65c](https://github.com/kurtosis-tech/kurtosis/commit/f09b65c36a0519757a089e629b0c299a3b41d466))
* support runtime values and ip addresses in exec and request ([#730](https://github.com/kurtosis-tech/kurtosis/issues/730)) ([f09b65c](https://github.com/kurtosis-tech/kurtosis/commit/f09b65c36a0519757a089e629b0c299a3b41d466))


### Bug Fixes

* re-enabled skipped tests ([5428e88](https://github.com/kurtosis-tech/kurtosis/commit/5428e8830926e043983e0992472a1508b3a37e4b))
* re-enabled skipped tests ([#726](https://github.com/kurtosis-tech/kurtosis/issues/726)) ([5428e88](https://github.com/kurtosis-tech/kurtosis/commit/5428e8830926e043983e0992472a1508b3a37e4b))

## [0.61.0](https://github.com/kurtosis-tech/kurtosis/compare/0.60.0...0.61.0) (2022-12-15)


### ⚠ BREAKING CHANGES

* exposed application protocol to users via starlark ([#649](https://github.com/kurtosis-tech/kurtosis/issues/649))
* added application protocol to starlark
* expose application protocol to kurtosis cli ([#641](https://github.com/kurtosis-tech/kurtosis/issues/641))
* exposed application protocol to cli
* expose app protocol to sdks ([#640](https://github.com/kurtosis-tech/kurtosis/issues/640))
* expose application protocol to users ([#703](https://github.com/kurtosis-tech/kurtosis/issues/703))

### Features

* Added `set_connection` instruction to configure a connection between two subnetworks ([#690](https://github.com/kurtosis-tech/kurtosis/issues/690)) ([6a8f6dd](https://github.com/kurtosis-tech/kurtosis/commit/6a8f6dd4a15b0cc4a6f05d1ed11653e109cb490b))
* Added `subnetwork` attribute to `ServiceConfig` to granularly add service to a partition when starting it ([#665](https://github.com/kurtosis-tech/kurtosis/issues/665)) ([9e1bc46](https://github.com/kurtosis-tech/kurtosis/commit/9e1bc46a6425ef8a6eb2ef3da99d73d3d9a16819))
* Added `UpdateService` function to `ServiceNetwork` to granularly update service partition once it is started ([#667](https://github.com/kurtosis-tech/kurtosis/issues/667)) ([3fdc6c3](https://github.com/kurtosis-tech/kurtosis/commit/3fdc6c327f80f7b76c364c0128c7626d12273925))
* added application protocol to starlark ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))
* expose app protocol to sdks ([#640](https://github.com/kurtosis-tech/kurtosis/issues/640)) ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))
* expose application protocol to kurtosis cli ([#641](https://github.com/kurtosis-tech/kurtosis/issues/641)) ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))
* expose application protocol to users ([#703](https://github.com/kurtosis-tech/kurtosis/issues/703)) ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))
* exposed application protocol to cli ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))
* exposed application protocol to users via starlark ([#649](https://github.com/kurtosis-tech/kurtosis/issues/649)) ([03b38b8](https://github.com/kurtosis-tech/kurtosis/commit/03b38b8e855341dbf20e2dbb2b3c383d56adaa7e))


### Bug Fixes

* made the filestore thread safe ([#710](https://github.com/kurtosis-tech/kurtosis/issues/710)) ([f297248](https://github.com/kurtosis-tech/kurtosis/commit/f29724852e42efb6999d95e9acf6888cc4e15df1))

## [0.60.0](https://github.com/kurtosis-tech/kurtosis/compare/0.59.3...0.60.0) (2022-12-15)


### ⚠ BREAKING CHANGES

* Updated 'KurtosisContext.GetServiceLogs' now receives a 'LogLineFilter' argument. Users will need to upgrade all the 'GetServiceLogs' call in order to pass the new 'LogLineFilter' argument ([#697](https://github.com/kurtosis-tech/kurtosis/issues/697))

### Features

* Added four constructors NewDoesContainTextLogLineFilter, NewDoesNotContainTextLogLineFilter, NewDoesContainMatchRegexLogLineFilter, and NewDoesNotContainMatchRegexLogLineFilter to create a LogLineFilter object, each constructor specifies the operator (text or regex expression) that will be used to filter the log lines. ([0956343](https://github.com/kurtosis-tech/kurtosis/commit/0956343ec3f2099fcf5109ef7f7760e535c90199))
* Added LogLineFilter object which can be used to filter the service logs when calling KurtosisContext.GetServiceLogs ([0956343](https://github.com/kurtosis-tech/kurtosis/commit/0956343ec3f2099fcf5109ef7f7760e535c90199))
* Updated 'KurtosisContext.GetServiceLogs' now receives a 'LogLineFilter' argument. Users will need to upgrade all the 'GetServiceLogs' call in order to pass the new 'LogLineFilter' argument ([#697](https://github.com/kurtosis-tech/kurtosis/issues/697)) ([0956343](https://github.com/kurtosis-tech/kurtosis/commit/0956343ec3f2099fcf5109ef7f7760e535c90199))


### Bug Fixes

* fix the update script ([5223aac](https://github.com/kurtosis-tech/kurtosis/commit/5223aac3c20c948a039d8299c25fd488b22e3c71))

## [0.59.3](https://github.com/kurtosis-tech/kurtosis/compare/0.59.2...0.59.3) (2022-12-14)


### Bug Fixes

* how we publish tags ([#686](https://github.com/kurtosis-tech/kurtosis/issues/686)) ([61d3234](https://github.com/kurtosis-tech/kurtosis/commit/61d3234a4edc2af6374d9b4fd6c502792c6909a2))

## [0.59.2](https://github.com/kurtosis-tech/kurtosis/compare/0.59.1...0.59.2) (2022-12-14)


### Features

* Added `SetConnection`, `UnsetConnection` and `SetDefaultConnection` to the `ServiceNetwork` to granularly update connections between two partitions ([#651](https://github.com/kurtosis-tech/kurtosis/issues/651)) ([40c8b56](https://github.com/kurtosis-tech/kurtosis/commit/40c8b564fec5f3fa30405c8e69c1395f224bc31b))

## 0.59.1

### Changes
- Cleaned up magic strings and updated protobuf to accept application protocol

## 0.59.0

### Breaking Changes
- Introduce `PortSpec` constructor to starlark to create port definitions

### Changes
- Added `does contain match regex` and `does not contain match regex` operators in the Kurtosis engine's server `GetServiceLogs` endpoint

## 0.58.2
### Features
- Made `args` optional for `run`
- Added metrics for `kurtosis run`

### Fixes
- Fix bug that panics APIC when `wait` assert fails
- Fixed the CLI output which could contain weird `%!p(MISSING)` when the output of a command was containing `%p` (or another Go formatting token)
- Cancel redundant runs of golang-ci-lint
- Fixed bug in installation of tab-completion

## 0.58.1
### Fixes
- Changed installation of tab-completion

## 0.58.0
### Breaking Changes
- Rename command from `get_value` to `request` command
- Change function signature of `wait` to take in a recipe, assertion and request interval/timeout
- Remove `extract` command
- Changed how `args` to `kurtosis run` are passed, they are passed as  second positional argument, instead of the `--args` flag
  - Users will have to start using `kurtosis run <script> <args>` without the `--arg` flag
  - If there are any scripts that depend on the `--args` flag, users should use the `args` arg instead
- Remove `define_fact` command

### Changes
- Add `extract` option to HTTP requests
- Prepared the Kurtosis engine server to do search in logs
- Adding `log line filters` parameter in the `GetServiceLogs` Kurtosis engine endpoint
- Made the test for `get_value` use the `jq` string extraction features
- Changed how `args` to `kurtosis run` are passed, they are passed as  second positional argument, instead of the `--args` flag
- Made `CLI` error if more arguments than expected are passed
- Added an advanced test for default_service_network.StartServices in preparation of changing a bit the logic
- Remove completion files
- CLI now prints to StdOut. It used to be printing most of its output to StdErr
- Remove build binary and completions directory from git

### Fixes
- Check an unchecked error in `CreateValue` in the `RunTimeValueStore`
- Appends `:latest` before checking for images without a version in `DockerManager.FetchImages`

### Features
- The CLI now displays the list of container images currently being downloaded and validated during the Starlark
validation step
- `exec` now returns the command output and code
- Added capability for container-engine to store optional application protocol for Kubernetes.
- Allow paths to `kurtosis.yml` to be run as Kurtosis packages

### Removals
- Remove facts engine and endpoints

## 0.57.8
### Features
- Added capability for container-engine to store optional application protocol for Docker.

## 0.57.7
### Changes
- Added automated installation of tab completion with brew installation.

### Fixes
- Fixed a bug which was happening on small terminal windows regarding the display of the progress bar and progress info

## 0.57.6
### Features
- The "Starlark code successfully executed" or "Error encountered running Starlark code" messages are now "Starlark
code successfully run in dry-run mode" and "Error encountered running Starlark code in dry-run mode" when Starlark is
run in dry-run mode (and without the "in dry-run mode" when the script is executed for real)\
- Added `RunStarlarkScriptBlocking`, `RunStarlarkPackageBlocking` and `RunStarlarkRemotePackageBlocking` functions
to the enclave context to facilitate automated testing in our current modules.

### Fixes
- Don't duplicate instruction position information in `store_service_files`
- Use constants instead of hardcoded string for validation errors

### Removals
- Remove stack trace from validation errors as it isn't used currently

### Changes
- Changed validation message from "Pre-validating" to "Validating"
- Disabled progress info in non-interactive terminals when running a Starlark Package

## 0.57.5
### Changes
- Replaced stack name with the stack file name in custom evaluation errors
- Replaced "internal ID" in the output message of `add_service` and `remove_service` instructions with "service GUID"

### Features
- Support public ports in Starlark to cover the NEAR usecase

### Fixes
- Corrected some old references to Starlark "modules"
- Fixed a typo where the CLI setup URL was redirecting to the CI setup
- Corrected almost all old references to `docs.kurtosistech.com`
- Changed the name from startosis to starlark in the `internal_testsuite` build script
- Fixed `internal-testsuites` omission during build time
- Fixed a bug related to omitting the `enclave ID` value when a function which filters modules is called

## 0.57.4
### Changes
- Simplified the API by removing the ServiceInfo struct that was adding unnecessary complexity.

## 0.57.3

### Changes
- Added exponential back-off and retries to `get_value`
- Removed `core-lib-documentation.md` and `engine-lib-documentation.md` in favour of the ones in the public docs repo

## 0.57.2

### Changes
- Added `startosis_add_service_with_empty_ports` Golang and Typescript internal tests

### Fixes
- Make validation more human-readable for missing docker images and instructions that depend on invalid service ids
- Fixed mismatch between `kurtosis enclave inspect` and `kurtosis enclave ls` displaying enclave creation time in different timezones

### Changes
- Make arg parsing errors more explicit on structs
- Updated Starlark section of core-lib-documentation.md to match the new streaming endpoints
- Updated `datastore-army-module` -> `datastore-army-package`
- Added `startosis_add_service_with_empty_ports` Golang and Typescript internal tests
- Removed `core-lib-documentation.md` and `engine-lib-documentation.md` in favour of the ones in the public docs repo

### Features
- Log file name and function like [filename.go:FunctionName()] while logging in `core` & `engine`
- Add artifact ID validation to Starlark commands
- Add IP address string replacement in `print` command
- All Kurtosis instructions now returns a simple but explicit output
- The object returned by Starlark's `run()` function is serialized as JSON and returned to the CLI output.
- Enforce `run(args)` for individual scripts

## 0.57.1

### Changes
- Added tab-completion (suggestions) to commands that require Service GUIDs, i.e.  `service shell` and `service logs` paths

## 0.57.0
### Breaking Changes
- Renamed `src_path` parameter in `read_file` to `src`
  - Users will have to upgrade their `read_file` calls to reflect this change

### Features
- Progress information (spinner, progress bar and quick progress message) is now printed by the CLI
- Instruction are now printed before the execution, and the associated result is printed once the execution is finished. This allows failed instruction to be printed before the error message is returned.

### Breaking changes
- Endpoints `ExecuteStartosisScript` and `ExecuteStartosisModule` were removed
- Endpoints `ExecuteKurtosisScript` was renamed `RunStarlarkScript` and `ExecuteKurtosisModule` was renamed `RunStarlarkPackage`

### Changes
- Starlark execution progress is now returned to the CLI via the KurtosisExecutionResponseLine stream
- Renamed `module` to `package` in the context of the Startosis engine

### Fixes
- Fixed the error message when the relative filename was incorrect in a Starlark import
- Fixed the error message when package name was incorrect
- Don't proceed with execution if there are validation errors in Starlark
- Made missing `run` method interpretation error more user friendly

## 0.56.0
### Breaking Changes
- Removed `module` key in the `kurtosis.yml` (formerly called `kurtosis.mod`) file to don't have nested keys
  - Users will have to update their `kurtosis.yml` to remove the key and move the `name` key in the root

## 0.55.1

### Changes
- Re-activate tests that had to be skipped because of the "Remove support for protobuf in Startosis" breaking change
- Renamed `input_args` to `args`. All Starlark packages should update `run(input_args)` to `run(args)`

## 0.55.0
### Fixes
- Fix failing documentation tests by linking to new domain in `cli`
- Fix failing `docs-checker` checks by pointing to `https://kurtosis-tech.github.io/kurtosis/` instead of `docs.kurtosistech.com`

### Breaking Changes
- Renamed `kurtosis.mod` file to `kurtosis.yml` this file extension enable syntax highlighting
  - Users will have to rename all theirs `kurtosis.mod` files

### Changes
- Made `run` an EngineConsumingKurtosisCommand, i.e. it automatically creates an engine if it doesn't exist
- Added serialized arguments to KurtosisInstruction API type such that the CLI can display executed instructions in a nicer way.

### Features
- Added one-off HTTP requests, `extract` and `assert`

## 0.54.1
### Fixes
- Fixes a bug where the CLI was returning 0 even when an error happened running a Kurtosis script

### Changes
- Small cleanup in `grpc_web_api_container_client` and `grpc_node_api_container_client`. They were implementing executeRemoteKurtosisModule unnecessarily

## 0.54.0
### Breaking Changes
- Renamed `kurtosis exec` to `kurtosis run` and `main in main.star` to `run in main.star`
  - Upgrade to the latest CLI, and use the `run` function instead
  - Upgrade existing modules to have `run` and not `main` in `main.star`

### Features
- Updated the CLI to consume the streaming endpoints to execute Startosis. Kurtosis Instructions are now returned live, but the script output is still printed at the end (until we have better formatting).
- Update integration tests to consume Startosis streaming endpoints

## 0.53.12
### Changes
- Changed occurrences of `[sS]tartosis` to `Starlark` in errors sent by the CLI and its long and short description
- Changed some logs and error messages inside core that which had references to Startosis to Starlark
- Allow `dicts` & `structs` to be passed to `render_templates.config.data`

## 0.53.11
### Changes
- Published the log-database HTTP port to the host machine

## 0.53.10
### Changes
- Add 2 endpoints to the APIC that streams the output of a Startosis script execution
- Changed the syntax of render_templates in Starlark

### Fixes
- Fixed the error that would happen if there was a missing `kurtosis.mod` file at the root of the module

## 0.53.9
### Fixes
- Renamed `artifact_uuid` to `artifact_id` and `src` to `src_path` in `upload_files` in Starlark

## 0.53.8

## 0.53.7
### Changes
- Modified the `EnclaveIdGenerator` now is a user defined type and can be initialized once because it contains a time-seed inside
- Simplify how the kurtosis instruction canonicalizer works. It now generates a single line canonicalized instruction, and indentation is performed at the CLI level using Bazel buildtools library.

### Fixes
- Fixed the `isEnclaveIdInUse` for the enclave validator, now uses on runtime for `is-key-in-map`

### Features
- Add the ability to execute remote modules using `EnclaveContext.ExecuteStartoisRemoteModule`
- Add the ability to execute remote module using cli `kurtosis exec github.com/author/module`

## 0.53.6

## 0.53.5
### Changes
- Error types in ExecuteStartosisResponse type is now a union type, to better represent they are exclusive and prepare for transition to streaming
- Update the KurtosisInstruction API type returned to the CLI. It now contains a combination of instruction position, the canonicalized instruction, and an optional instruction result
- Renamed `store_files_from_service` to `store_service_files`
- Slightly update the way script output information are passed from the Startosis engine back the API container main class. This is a step to prepare for streaming this output all the way back the CLI.
- Removed `load` statement in favour of `import_module`. Calling load will now throw an InterpretationError
- Refactored startosis tests to enable parallel execution of tests

## 0.53.4

## 0.53.3
### Fixes
- Fixed a bug with dumping enclave logs during the CI run

### Features
- Log that the module is being compressed & uploaded during `kurtosis exec`
- Added `file_system_path_arg` in the CLI which provides validation and tab auto-completion for filepath, dirpath, or both kind of arguments
- Added tab-auto-complete for the `script-or-module-path` argument in `kurtosis exec` CLI command

### Changes
- `print()` is now a regular instructions like others, and it takes effect at execution time (used to be during interpretation)
- Added `import_module` startosis builtin to replace `load`. Load is now deprecated. It can still be used but it will log a warning. It will be entirely removed in a future PR
- Added exhaustive struct linting and brought code base into exhaustive struct compliance
- Temporarily disable enclave dump for k8s in CircleCI until we fix issue #407
- Small cleanup to kurtosis instruction classes. It now uses a pointer to the position object.

### Fixes
- Renamed `cmd_args` and `entrypoint_args` inside `config` inside `add_service` to `cmd` and `entrypoint`

### Breaking Changes
- Renamed `cmd_args` and `entrypoint_args` inside `config` inside `add_service` to `cmd` and `entrypoint`
  - Users will have to replace their use of `cmd_args` and `entry_point_args` to the above inside their Starlark modules

## 0.53.2
### Features
- Make facts referencable on `add_service`
- Added a new log line for printing the `created enclave ID` just when this is created in `kurtosis exec` and `kurtosis module exec` commands

## 0.53.1
### Features
- Added random enclave ID generation in `EnclaveManager.CreateEnclave()` when an empty enclave ID is provided
- Added the `created enclave` spotlight message when a new enclave is created from the CLI (currently with the `enclave add`, `module exec` and `exec` commands)

### Changes
- Moved the enclave ID auto generation and validation from the CLI to the engine's server which will catch all the presents and future use cases

### Fixes
- Fixed a bug where we had renamed `container_image_name` inside the proto definition to `image`
- Fix a test that dependent on an old on existent Starlark module

## 0.53.0
### Features
- Made `render_templates`, `upload_files`, `store_Files_from_service` accept `artifact_uuid` and
return `artifact_uuid` during interpretation time
- Moved `kurtosis startosis exec` to `kurtosis exec`

### Breaking Features
- Moved `kurtosis startosis exec` to `kurtosis exec`
  - Users now need to use the new command to launch Starlark programs

### Fixes
- Fixed building kurtosis by adding a conditional to build.sh to ignore startosis folder under internal_testsuites

## 0.52.5
### Fixes
- Renamed `files_artifact_mount_dirpaths` to just `files`

## 0.52.4
### Features
- Added the enclave's creation time info which can be obtained through the `enclave ls` and the `enclave inspect` commands

### Fixes
- Smoothened the experience `used_ports` -> `ports`, `container_image_name` -> `name`, `service_config` -> `config`

## 0.52.3
### Changes
- Cleanup Startosis interpreter predeclared functions

<<<<<<< HEAD:docs/changelog.md
# 0.52.2
### Fixes
- Fixed TestValidUrls so that it checks for the correct http return code
||||||| e322558df:docs/changelog.md
# 0.52.2
=======
## 0.52.2
>>>>>>> master:CHANGELOG.md

## 0.52.1
### Features
- Add `wait` and `define` command in Startosis
- Added `not found service GUIDs information` in `KurtosisContext.GetServiceLogs` method
- Added a warning message in `service logs` CLI command when the request service GUID is not found in the logs database
- Added ip address replacement in the JSON for `render_template` instruction

### Changes
- `kurtosis_instruction.String()` now returns a single line version of the instruction for more concise logging

### Fixes
- Fixes a bug where we'd propagate a nil error
- Adds validation for `service_name` in `store_files_from_service`
- Fixes a bug where typescript (jest) unit tests do not correctly wait for grpc services to become available
- Fixed a panic that would happen cause of a `nil` error being returned

## 0.52.0
### Breaking Changes
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` engine's endpoints, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `GetUserServiceLogsArgs` engine's proto file
  - Users should have to add this param in all the `GetUserServiceLogs` calls
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` methods in `KurtosisContext`, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `KurtosisContext.GetUserServiceLogs`
  - Users will have to addition this new parameter on every call

### Changes
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON
- Moved `read_file` to be a simple Startosis builtin in place of a Kurtosis instruction

## 0.51.13
### Fixes
- Set `entrypoint` and `cmd_args` to `nil` if not specified instead of empty array

## 0.51.12
### Features
- Added an optional `--dry-run` flag to the `startosis exec` (defaulting to false) command which prints the list of Kurtosis instruction without executing any. When `--dry-run` is set to false, the list of Kurtosis instructions is printed to the output of CLI after being executed.

## 0.51.11
### Features
- Improve how kurtosis instructions are canonicalized with a universal canonicalizer. Each instruction is now printed on multiple lines with a comment pointing the to position in the source code.
- Support `private_ip_address_placeholder` to be passed in `config` for `add_service` in Starlark

### Changes
- Updated how we generate the canonical string for Kurtosis `upload_files` instruction

## 0.51.10
### Changes
- Added Starlark `proto` module, such that you can now do `proto.has(msg, "field_name")` in Startosis to differentiate between when a field is set to its default value and when it is unset (the field has to be marked as optional) in the proto file though.

## 0.51.9
### Features
- Implemented the new `StreamUserServiceLogs` endpoint in the Kurtosis engine server
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Golang library
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Typescript library
- Added the `StreamUserServiceLogs` method in Loki logs database client
- Added `stream-logs` test in Golang and Typescript `internal-testsuites`
- Added `service.GUID` field in `Service.Ctx` in the Kurtosis SDK

### Changes
- Updated the CLI `service logs` command in order to use the new `KurtosisContext.StreamUserServiceLogs` when user requested to follow logs
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON

## 0.51.8
### Features
- Added exec and HTTP request facts
- Prints out the instruction line, col & filename in the execution error
- Prints out the instruction line, col & filename in the validation error
- Added `remove_service` to Startosis

### Fixes
- Fixed nil accesses on Fact Engine

### Changes
- Add more integration tests for Kurtosis modules with input and output types

## 0.51.7
### Fixes
- Fixed instruction position to work with nested functions

### Features
- Instruction position now contains the filename too

## 0.51.6
### Features
- Added an `import_types` Starlark instruction to read types from a .proto file inside a module
- Added the `time` module for Starlark to the interpreter
- Added the ability for a Starlark module to take input args when a `ModuleInput` in the module `types.proto` file

## 0.51.5
### Fixes
- Testsuite CircleCI jobs also short-circuit if the only changes are to docs, to prevent them failing due to no CLI artifact

## 0.51.4
### Fixes
- Fixed a bug in `GetLogsCollector` that was failing when there is an old logs collector container running that doesn't publish the TCP port
- Add missing bindings to Kubernetes gateway

### Changes
- Adding/removing methods from `.proto` files will now be compile errors in Go code, rather than failing at runtime
- Consolidated the core & engine Protobuf regeneration scripts into a single one

### Features
- Validate service IDs on Startosis commands

## 0.51.3
### Fixes
- Added `protoc` install step to the `publish_api_container_server_image` CircleCI task

## 0.51.2
### Features
- Added a `render_templates` command to Startosis
- Implemented backend for facts engine
- Added a `proto_file_store` in charge of compiling Startosis module's .proto file on the fly and storing their FileDescriptorSet in memory

### Changes
- Simplified own-version constant generation by checking in `kurtosis_version` directory

## 0.51.1
- Added an `exec` command to Startosis
- Added a `store_files_from_service` command to Startosis
- Added the ability to pass `files` to the service config
- Added a `read_file` command to Startosis
- Added the ability to execute local modules in Startosis

### Changes
- Fixed a typo in a filename

### Fixes
- Fixed a bug in exec where we'd propagate a `nil` error
- Made the `startosis_module_test` in js & golang deterministic and avoid race conditions during parallel runs

### Removals
- Removed  stale `scripts/run-pre-release-scripts` which isn't used anywhere and is invalid.

## 0.51.0
### Breaking Changes
- Updated `kurtosisBackend.CreateLogsCollector` method in `container-engine-lib`, added the `logsCollectorTcpPortNumber` parameter
  - Users will need to update all the `kurtosisBackend.CreateLogsCollector` setting the logs collector `TCP` port number

### Features
- Added `KurtosisContext.GetUserServiceLogs` method in `golang` and `typescript` api libraries
- Added the public documentation for the new `KurtosisContext.GetUserServiceLogs` method
- Added `GetUserServiceLogs` in Kurtosis engine gateway
- Implemented IP address references for services
- Added the `defaultTcpLogsCollectorPortNum` with `9713` value in `EngineManager`
- Added the `LogsCollectorAvailabilityChecker` interface

### Changes
- Add back old enclave continuity test
- Updated the `FluentbitAvailabilityChecker` constructor now it also receives the IP address as a parameter instead of using `localhost`
- Published the `FluentbitAvailabilityChecker` constructor for using it during starting modules and user services
- Refactored `service logs` Kurtosis CLI command in order to get the user service logs from the `logs database` (implemented in Docker cluster so far)

## 0.50.2
### Fixes
- Fixes how the push cli artifacts & publish engine runs by generating kurtosis_version before hand

## 0.50.1
### Fixes
- Fix generate scripts to take passed version on release

## 0.50.0
### Features
- Created new engine's endpoint `GetUserServiceLogs` for consuming user service container logs from the logs database server
- Added `LogsDatabaseClient` interface for defining the behaviour for consuming logs from the centralized logs database
- Added `LokiLogsDatabaseClient` which implements `LogsDatabaseClient` for consuming logs from a Loki's server
- Added `KurtosisBackendLogsClient` which implements `LogsDatabaseClient` for consuming user service container logs using `KurtosisBackend`
- Created the `LogsDatabase` object in `container-engine-lib`
- Created the `LogsCollector` object in `container-engine-lib`
- Added `LogsDatabase` CRUD methods in `Docker` Kurtosis backend
- Added `LogsCollector` CRUD methods in `Docker` Kurtosis backend
- Added `ServiceNetwork` (interface), `DefaultServiceNetwork` and `MockServiceNetwork`

### Breaking Changes
- Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
  - Users will need to update all the `CreateEngine` calls removing this parameter
- Updated `NewEngineServerArgs`,  `LaunchWithDefaultVersion` and `LaunchWithCustomVersion` methods in `engine_server_launcher` removed the `logsCollectorHttpPortNumber` parameter
  - Users will need to update these method calls removing this parameter

### Changes
- Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`
- Made some changes to the implementation of the module manager based on some PR comments by Kevin

### Features
- Implement Startosis add_service image pull validation
- Startosis scripts can now be run from the CLI: `kurtosis startosis exec path/to/script/file --enclave-id <ENCLAVE_ID>`
- Implemented Startosis load method to load from Github repositories

### Fixes
- Fix IP address placeholder injected by default in Startosis instructions. It used to be empty, which is invalid now
it is set to `KURTOSIS_IP_ADDR_PLACEHOLDER`
- Fix enclave inspect CLI command error when there are additional port bindings
- Fix a stale message the run-all-test-against-latest-code script
- Fix bug that creates database while running local unit tests
- Manually truncate string instead of using `k8s.io/utils/strings`

### Removals
- Removes version constants within launchers and cli in favor of centralized generated version constant
- Removes remote-docker-setup from the `build_cli` job in Circle

## 0.49.9
### Features
- Implement Startosis add_service method
- Enable linter on Startosis codebase

## 0.49.8
### Changes
- Added a linter
- Made changes based on the linters output
- Made the `discord` command a LowLevelKurtosisCommand instead of an EngineConsumingKurtosisCommand

### Features
- API container now saves free IPs on a local database

### Fixes
- Fix go.mod for commons & cli to reflect monorepo and replaced imports with write package name
- Move linter core/server linter config to within core/server

## 0.49.7
### Features
- Implement skeleton for the Startosis engine

### Fixes
- Fixed a message that referred to an old repo.

### Changes
- Added `cli` to the monorepo

## 0.49.6
### Fixes
- Fixed a bug where engine launcher would try to launch older docker image `kurtosistech/kurtosis-engine-server`.

## 0.49.5
### Changes
- Added `kurtosis-engine-server` to the monorepo
- Merged the `kurtosis-engine-sdk` & `kurtosis-core-sdk`

### Removals
- Remove unused variables from Docker Kurtosis backend

## 0.49.4
### Fixes
- Fix historical changelog for `kurtosis-core`
- Don't check for grpc proxy to be available

## 0.49.3
### Fixes
- Fix typescript package releases

## 0.49.2
### Removals
- Remove envoy proxy from docker image. No envoy proxy is being run anymore, effectively removing HTTP1.

### Changes
- Added `kurtosis-core` to the monorepo

### Fixes
- Fixed circle to not docs check on merge

## 0.49.1
### Fixes
- Attempting to fix the release version
### Changes
- Added container-engine-lib

## 0.49.0
### Changes
- This version is a dummy version to set the minimum. We pick a version greater than the current version of the CLI (0.29.1).

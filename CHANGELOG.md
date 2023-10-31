# Changelog

## [0.85.2](https://github.com/kurtosis-tech/kurtosis/compare/0.85.1...0.85.2) (2023-10-31)


### Bug Fixes

* block local absolute locators ([#1659](https://github.com/kurtosis-tech/kurtosis/issues/1659)) ([a4daeb3](https://github.com/kurtosis-tech/kurtosis/commit/a4daeb3437219245d5b04a15f28b6addae1c29b6)), closes [#1637](https://github.com/kurtosis-tech/kurtosis/issues/1637)
* use full path for running black to allow older versions of docker ([#1666](https://github.com/kurtosis-tech/kurtosis/issues/1666)) ([fdcd3d9](https://github.com/kurtosis-tech/kurtosis/commit/fdcd3d94086365e62bab07cf34a91df2fa5bff73))

## [0.85.1](https://github.com/kurtosis-tech/kurtosis/compare/0.85.0...0.85.1) (2023-10-31)


### Features

* initial new enclave manager ([#1603](https://github.com/kurtosis-tech/kurtosis/issues/1603)) ([9944658](https://github.com/kurtosis-tech/kurtosis/commit/9944658f5036d45dde64721e9a958322e17b9d5b))
* warn users of diff architecture image running ([#1649](https://github.com/kurtosis-tech/kurtosis/issues/1649)) ([77f2f69](https://github.com/kurtosis-tech/kurtosis/commit/77f2f694210e35d98b37b31396596791a2a2d0c7))

## [0.85.0](https://github.com/kurtosis-tech/kurtosis/compare/0.84.13...0.85.0) (2023-10-30)


### ⚠ BREAKING CHANGES

* protobuf definitions for more idiomatic SDKs ([#1586](https://github.com/kurtosis-tech/kurtosis/issues/1586))

### Features

* Add cli argument to control image download ([#1495](https://github.com/kurtosis-tech/kurtosis/issues/1495)) ([f210a76](https://github.com/kurtosis-tech/kurtosis/commit/f210a7604a283d014d79eff109654486c0b7cc83))


### Bug Fixes

* run_sh doesn't remove new lines from input ([#1642](https://github.com/kurtosis-tech/kurtosis/issues/1642)) ([a969dff](https://github.com/kurtosis-tech/kurtosis/commit/a969dffd1902952c4500c4f329480909e3f81dfd))


### Code Refactoring

* protobuf definitions for more idiomatic SDKs ([#1586](https://github.com/kurtosis-tech/kurtosis/issues/1586)) ([e7ab58a](https://github.com/kurtosis-tech/kurtosis/commit/e7ab58a1d2a286fcfb9af35e01997c2e05f7a107)), closes [#843](https://github.com/kurtosis-tech/kurtosis/issues/843)

## [0.84.13](https://github.com/kurtosis-tech/kurtosis/compare/0.84.12...0.84.13) (2023-10-25)


### Features

* user-configurable labels (in ServiceConfig type) for Docker containers and k8s pods ([#1604](https://github.com/kurtosis-tech/kurtosis/issues/1604)) ([e98cdf6](https://github.com/kurtosis-tech/kurtosis/commit/e98cdf6874b610f158a0ff798a01cf9a1b70d183))


### Bug Fixes

* name temporary python script for run_python with suitable name ([#1616](https://github.com/kurtosis-tech/kurtosis/issues/1616)) ([88edb39](https://github.com/kurtosis-tech/kurtosis/commit/88edb39c8f424d5f6b2126739948206ce5829e98))

## [0.84.12](https://github.com/kurtosis-tech/kurtosis/compare/0.84.11...0.84.12) (2023-10-25)


### Features

* kurtosis run command now accepts URLs with the 'args-file' argument  ([#1607](https://github.com/kurtosis-tech/kurtosis/issues/1607)) ([ec32d0f](https://github.com/kurtosis-tech/kurtosis/commit/ec32d0f48f0a1cd76e26e4fdeecc75e7c1a31929)), closes [#1596](https://github.com/kurtosis-tech/kurtosis/issues/1596)
* Product and subscription added to the get payment config response ([#1606](https://github.com/kurtosis-tech/kurtosis/issues/1606)) ([0d10726](https://github.com/kurtosis-tech/kurtosis/commit/0d107261422ad918b4a5dbc5dbbb35c8d555d4c5))


### Bug Fixes

* add a debug line for the exact command run by lint ([#1615](https://github.com/kurtosis-tech/kurtosis/issues/1615)) ([3fa6d2f](https://github.com/kurtosis-tech/kurtosis/commit/3fa6d2f62b301f97e7ae7ef50b9abe460e7cc283))
* handle error and fix rendering bug ([#1617](https://github.com/kurtosis-tech/kurtosis/issues/1617)) ([825fd22](https://github.com/kurtosis-tech/kurtosis/commit/825fd2238601f7a95c97f3a695773d3a9c234c49))

## [0.84.11](https://github.com/kurtosis-tech/kurtosis/compare/0.84.10...0.84.11) (2023-10-24)


### Features

* add full story script ([#1610](https://github.com/kurtosis-tech/kurtosis/issues/1610)) ([de10c7b](https://github.com/kurtosis-tech/kurtosis/commit/de10c7bab36c0c7ee1bea99d181d93134e138d04))
* allow for named artifact creation in run_sh and run_python ([#1608](https://github.com/kurtosis-tech/kurtosis/issues/1608)) ([1a9d953](https://github.com/kurtosis-tech/kurtosis/commit/1a9d953bb26643c7b0effcf761de49bdb735a0ec)), closes [#1581](https://github.com/kurtosis-tech/kurtosis/issues/1581)
* disable smooth scrolling for logs and default to select restart services ([#1612](https://github.com/kurtosis-tech/kurtosis/issues/1612)) ([2ee86c4](https://github.com/kurtosis-tech/kurtosis/commit/2ee86c419e081996711ff54e009e8333df28839c))


### Bug Fixes

* clean em api get service logs streaming logic ([#1589](https://github.com/kurtosis-tech/kurtosis/issues/1589)) ([f8d8bda](https://github.com/kurtosis-tech/kurtosis/commit/f8d8bda8783995d3c22801fea6586c4af2fc1677))
* show container status instead of service status  in UI ([#1567](https://github.com/kurtosis-tech/kurtosis/issues/1567)) ([4b75980](https://github.com/kurtosis-tech/kurtosis/commit/4b759806cd522b03fa7eadbfc83f952ded6b1b20))

## [0.84.10](https://github.com/kurtosis-tech/kurtosis/compare/0.84.9...0.84.10) (2023-10-23)


### Bug Fixes

* bug in portal forwarding via run ([#1598](https://github.com/kurtosis-tech/kurtosis/issues/1598)) ([bf534c3](https://github.com/kurtosis-tech/kurtosis/commit/bf534c35055f4ec3e19cc1f1e2e32e8d29e61b5a))

## [0.84.9](https://github.com/kurtosis-tech/kurtosis/compare/0.84.8...0.84.9) (2023-10-19)


### Features

* Cloud backend method to refresh the default payment method ([#1569](https://github.com/kurtosis-tech/kurtosis/issues/1569)) ([9f3d808](https://github.com/kurtosis-tech/kurtosis/commit/9f3d808264fa1d83d6fd0d2ea3dde3b5d3f9e48e))
* get package on demand ([#1590](https://github.com/kurtosis-tech/kurtosis/issues/1590)) ([0af4086](https://github.com/kurtosis-tech/kurtosis/commit/0af4086da6887aa80d7a1b06d232254178a022e7))

## [0.84.8](https://github.com/kurtosis-tech/kurtosis/compare/0.84.7...0.84.8) (2023-10-17)


### Features

* kurtosis package init command ([#1547](https://github.com/kurtosis-tech/kurtosis/issues/1547)) ([6411c8f](https://github.com/kurtosis-tech/kurtosis/commit/6411c8f8b8f2ed3737d04c6d8a7a0938f7486aa3))


### Bug Fixes

* correct the link to kurtosis upgrade docs ([#1574](https://github.com/kurtosis-tech/kurtosis/issues/1574)) ([11d1dba](https://github.com/kurtosis-tech/kurtosis/commit/11d1dba7541fc87fdf0e6bee3efe345edd732c23))
* error clearly if there are no nodes on the Kubernetes cluster ([#1553](https://github.com/kurtosis-tech/kurtosis/issues/1553)) ([77f9ad4](https://github.com/kurtosis-tech/kurtosis/commit/77f9ad42ba18f00faea5937bf3a971056ea8720b))

## [0.84.7](https://github.com/kurtosis-tech/kurtosis/compare/0.84.6...0.84.7) (2023-10-16)


### Features

* Add create enclave utils to SDK ([#1550](https://github.com/kurtosis-tech/kurtosis/issues/1550)) ([eb952bb](https://github.com/kurtosis-tech/kurtosis/commit/eb952bb9d00ff30adeb3e78a71f39d6f546dd180))
* provide granular progress of starlark package run ([#1548](https://github.com/kurtosis-tech/kurtosis/issues/1548)) ([8b20031](https://github.com/kurtosis-tech/kurtosis/commit/8b2003109f426ab3ba6498b63eb37dad4c697e40))
* rename kurtosis context "switch" to "set" ([#1537](https://github.com/kurtosis-tech/kurtosis/issues/1537)) ([ccff275](https://github.com/kurtosis-tech/kurtosis/commit/ccff2756b53e84516376c41ff1a36958b072acf3))


### Bug Fixes

* propagate unexpected test errors via the test framework ([#1559](https://github.com/kurtosis-tech/kurtosis/issues/1559)) ([c463ae2](https://github.com/kurtosis-tech/kurtosis/commit/c463ae278b0d8846edcbc248784f56fdb74ad5be))
* show container status instead of service status in enclave inspect ([#1560](https://github.com/kurtosis-tech/kurtosis/issues/1560)) ([3e1208b](https://github.com/kurtosis-tech/kurtosis/commit/3e1208bc9340302db49a041fc93b1e2d565e6abc)), closes [#1351](https://github.com/kurtosis-tech/kurtosis/issues/1351)

## [0.84.6](https://github.com/kurtosis-tech/kurtosis/compare/0.84.5...0.84.6) (2023-10-13)


### Features

* Unused images are cleaned even without -a flag ([#1551](https://github.com/kurtosis-tech/kurtosis/issues/1551)) ([e1317aa](https://github.com/kurtosis-tech/kurtosis/commit/e1317aaa0853943d73234dde344fa5102ab41bd8)), closes [#1523](https://github.com/kurtosis-tech/kurtosis/issues/1523)

## [0.84.5](https://github.com/kurtosis-tech/kurtosis/compare/0.84.4...0.84.5) (2023-10-12)


### Features

* highlight the active cluster in kurtosis cluster ls ([#1514](https://github.com/kurtosis-tech/kurtosis/issues/1514)) ([67e0111](https://github.com/kurtosis-tech/kurtosis/commit/67e0111af7483efdab14743f5e10897054db96a2))
* local replace package dependency ([#1521](https://github.com/kurtosis-tech/kurtosis/issues/1521)) ([d5e3126](https://github.com/kurtosis-tech/kurtosis/commit/d5e3126900f1a16523a3c8ba33b25c3a7bed6e0d))
* manage script return value ([#1546](https://github.com/kurtosis-tech/kurtosis/issues/1546)) ([a53508f](https://github.com/kurtosis-tech/kurtosis/commit/a53508f825985a26e306e10038305798f3e3ce4d))


### Bug Fixes

* run package bug ([#1539](https://github.com/kurtosis-tech/kurtosis/issues/1539)) ([1f5380a](https://github.com/kurtosis-tech/kurtosis/commit/1f5380afeb91c3dfe8b365b7752494f5444376e7)), closes [#1501](https://github.com/kurtosis-tech/kurtosis/issues/1501) [#1479](https://github.com/kurtosis-tech/kurtosis/issues/1479)

## [0.84.4](https://github.com/kurtosis-tech/kurtosis/compare/0.84.3...0.84.4) (2023-10-10)


### Features

* Always keep latest released version of Kurtosis images ([#1473](https://github.com/kurtosis-tech/kurtosis/issues/1473)) ([7fbdfd0](https://github.com/kurtosis-tech/kurtosis/commit/7fbdfd0abbf13232357322e8fe51ef6b36d082a3))
* make clean -a remove all logs ([#1517](https://github.com/kurtosis-tech/kurtosis/issues/1517)) ([3ec7d88](https://github.com/kurtosis-tech/kurtosis/commit/3ec7d88a3dcec7a33ac5003c7bb167fe5c4805b9))


### Bug Fixes

* check docker engine is prior to linting and give a useful error when it is not ([#1506](https://github.com/kurtosis-tech/kurtosis/issues/1506)) ([542d435](https://github.com/kurtosis-tech/kurtosis/commit/542d4351fd75391adea537a260bef0aaa7d98eb8))
* set parallelism to 4 when its passed as 0 ([#1502](https://github.com/kurtosis-tech/kurtosis/issues/1502)) ([4af67d5](https://github.com/kurtosis-tech/kurtosis/commit/4af67d5d12af30919afcdf8701432ab6ee92a4ca))

## [0.84.3](https://github.com/kurtosis-tech/kurtosis/compare/0.84.2...0.84.3) (2023-10-09)


### Features

* regular replace package dependency and replace package with no-main-branch ([#1481](https://github.com/kurtosis-tech/kurtosis/issues/1481)) ([bec49ac](https://github.com/kurtosis-tech/kurtosis/commit/bec49ac496d763d4a3002433274d684d7fc06a62))
* remove logs on enclave rm and clean -a ([#1489](https://github.com/kurtosis-tech/kurtosis/issues/1489)) ([9ea344e](https://github.com/kurtosis-tech/kurtosis/commit/9ea344ededfaac909342df32dddf23757c8e873d))


### Bug Fixes

* Add new line while inspecting file contents ([#1477](https://github.com/kurtosis-tech/kurtosis/issues/1477)) ([545aa53](https://github.com/kurtosis-tech/kurtosis/commit/545aa53d6e86b1cb2be6e8f118d34380d656f583))
* improve absolute locator checks ([#1498](https://github.com/kurtosis-tech/kurtosis/issues/1498)) ([cda001d](https://github.com/kurtosis-tech/kurtosis/commit/cda001d08b9d332a18d661ac8eae2c081511c538))
* kurtosis web cmd work for remote context ([#1486](https://github.com/kurtosis-tech/kurtosis/issues/1486)) ([8d8634c](https://github.com/kurtosis-tech/kurtosis/commit/8d8634c6d9b6ee7925346b79eef82ecf5e5b40da))
* make vector use ISO week time ([#1497](https://github.com/kurtosis-tech/kurtosis/issues/1497)) ([e6d1f5e](https://github.com/kurtosis-tech/kurtosis/commit/e6d1f5e536e84a5121103a78117d37e9baf5ca4a))
* replace duplicate log files with symlinks ([#1472](https://github.com/kurtosis-tech/kurtosis/issues/1472)) ([57da901](https://github.com/kurtosis-tech/kurtosis/commit/57da901da2bc76fc345b4060225db39a550de023))

## [0.84.2](https://github.com/kurtosis-tech/kurtosis/compare/0.84.1...0.84.2) (2023-10-05)


### Features

* Edit Enclave ([#1478](https://github.com/kurtosis-tech/kurtosis/issues/1478)) ([d11736a](https://github.com/kurtosis-tech/kurtosis/commit/d11736a8a28bbda44b7499bd9977b4f00b1bea74))
* Get Starlark Run APIC endpoint ([#1455](https://github.com/kurtosis-tech/kurtosis/issues/1455)) ([503cb8d](https://github.com/kurtosis-tech/kurtosis/commit/503cb8d5ad781b51e96524d8fa7068478370e5dd))

## [0.84.1](https://github.com/kurtosis-tech/kurtosis/compare/0.84.0...0.84.1) (2023-10-04)


### Bug Fixes

* autoscroll ([#1471](https://github.com/kurtosis-tech/kurtosis/issues/1471)) ([9948fad](https://github.com/kurtosis-tech/kurtosis/commit/9948fad01bd884cf75fbc832dd41516bc50cd6a6))
* bug where we passed cloud user id for cloud instance id ([#1465](https://github.com/kurtosis-tech/kurtosis/issues/1465)) ([65b749c](https://github.com/kurtosis-tech/kurtosis/commit/65b749cd2e8c612a6ecabbe72eb1c29e9c53ced3))

## [0.84.0](https://github.com/kurtosis-tech/kurtosis/compare/0.83.16...0.84.0) (2023-10-03)


### ⚠ BREAKING CHANGES

* block 'local absolute locators', users should replace' local absolute locators' with 'relative locators' ([#1446](https://github.com/kurtosis-tech/kurtosis/issues/1446))
* move run metrics to the APIC & refactor SDK ([#1456](https://github.com/kurtosis-tech/kurtosis/issues/1456))

### Features

* block 'local absolute locators', users should replace' local absolute locators' with 'relative locators' ([#1446](https://github.com/kurtosis-tech/kurtosis/issues/1446)) ([27ded02](https://github.com/kurtosis-tech/kurtosis/commit/27ded02f79b71998f378ea3728ae8c93358738c7))


### Bug Fixes

* add navigation to kurtosis enclave manager  ([#1458](https://github.com/kurtosis-tech/kurtosis/issues/1458)) ([f27a00a](https://github.com/kurtosis-tech/kurtosis/commit/f27a00aced78d48c04961c7238907e9c8a3f0261))


### Code Refactoring

* move run metrics to the APIC & refactor SDK ([#1456](https://github.com/kurtosis-tech/kurtosis/issues/1456)) ([2a6c908](https://github.com/kurtosis-tech/kurtosis/commit/2a6c9080f470385f6d0f4ee5c54ab73a25d97c5f))

## [0.83.16](https://github.com/kurtosis-tech/kurtosis/compare/0.83.15...0.83.16) (2023-10-02)


### Features

* add font ([#1454](https://github.com/kurtosis-tech/kurtosis/issues/1454)) ([75ce332](https://github.com/kurtosis-tech/kurtosis/commit/75ce3323c2b593d3f73f47866307e3476c6633d5)), closes [#1386](https://github.com/kurtosis-tech/kurtosis/issues/1386)
* added --args-file to Kurtosis run ([#1451](https://github.com/kurtosis-tech/kurtosis/issues/1451)) ([fdc6220](https://github.com/kurtosis-tech/kurtosis/commit/fdc622074b9ddb78f0265885c57208ac1e28fc9d)), closes [#1112](https://github.com/kurtosis-tech/kurtosis/issues/1112)


### Bug Fixes

* Remove mouse wheel capture ([#1452](https://github.com/kurtosis-tech/kurtosis/issues/1452)) ([2d35d77](https://github.com/kurtosis-tech/kurtosis/commit/2d35d7731019018051921b9b97735ce90277a68c)), closes [#1438](https://github.com/kurtosis-tech/kurtosis/issues/1438)

## [0.83.15](https://github.com/kurtosis-tech/kurtosis/compare/0.83.14...0.83.15) (2023-10-02)


### Features

* Add product area to the bug report template ([#1441](https://github.com/kurtosis-tech/kurtosis/issues/1441)) ([6d07ed6](https://github.com/kurtosis-tech/kurtosis/commit/6d07ed68005bdaf785328ef8f48d9b6560d185c9))
* added a tool tip to show users new logs are present ([#1444](https://github.com/kurtosis-tech/kurtosis/issues/1444)) ([82ce14b](https://github.com/kurtosis-tech/kurtosis/commit/82ce14ba2ce667668350ec173e45594d7dbe9089))


### Bug Fixes

* relative locators for read_file and upload_files instructions ([#1427](https://github.com/kurtosis-tech/kurtosis/issues/1427)) ([e5d2c54](https://github.com/kurtosis-tech/kurtosis/commit/e5d2c5462471c505853ae2d02d8260c604fbcf38)), closes [#1412](https://github.com/kurtosis-tech/kurtosis/issues/1412)

## [0.83.14](https://github.com/kurtosis-tech/kurtosis/compare/0.83.13...0.83.14) (2023-09-29)


### Bug Fixes

* scroll tracking experience is improved. ([#1429](https://github.com/kurtosis-tech/kurtosis/issues/1429)) ([0572a5c](https://github.com/kurtosis-tech/kurtosis/commit/0572a5c6026a15096e38ddd22d709dc7b6be4edc))

## [0.83.13](https://github.com/kurtosis-tech/kurtosis/compare/0.83.12...0.83.13) (2023-09-28)


### Features

* add extra verification when deleting prod enclave. ([#1404](https://github.com/kurtosis-tech/kurtosis/issues/1404)) ([6e3ea07](https://github.com/kurtosis-tech/kurtosis/commit/6e3ea07437368f1fdfbbc32688d8434b223c4ef1))
* Get user payment config method definition ([#1374](https://github.com/kurtosis-tech/kurtosis/issues/1374)) ([c52cb97](https://github.com/kurtosis-tech/kurtosis/commit/c52cb97e7427c3796b9eedbf1fb115320060c664))


### Bug Fixes

* build ([#1432](https://github.com/kurtosis-tech/kurtosis/issues/1432)) ([4e7b618](https://github.com/kurtosis-tech/kurtosis/commit/4e7b6187944972fd12191f97a8064ed62c6db183)), closes [#1425](https://github.com/kurtosis-tech/kurtosis/issues/1425)

## [0.83.12](https://github.com/kurtosis-tech/kurtosis/compare/0.83.11...0.83.12) (2023-09-28)


### Bug Fixes

* stale data in code editor ([#1421](https://github.com/kurtosis-tech/kurtosis/issues/1421)) ([d58ca3f](https://github.com/kurtosis-tech/kurtosis/commit/d58ca3f6fcb8a7cbf996c49cea53bf9b71c4987a))

## [0.83.11](https://github.com/kurtosis-tech/kurtosis/compare/0.83.10...0.83.11) (2023-09-27)


### Bug Fixes

* make linter work with individual files ([#1378](https://github.com/kurtosis-tech/kurtosis/issues/1378)) ([edcd8c8](https://github.com/kurtosis-tech/kurtosis/commit/edcd8c8c9ebc0a46612739fdc97bee772e011f12))

## [0.83.10](https://github.com/kurtosis-tech/kurtosis/compare/0.83.9...0.83.10) (2023-09-27)


### Features

* disable scrollbar, remove line highlighting, set background color ([#1408](https://github.com/kurtosis-tech/kurtosis/issues/1408)) ([1ffdf10](https://github.com/kurtosis-tech/kurtosis/commit/1ffdf10e985b48e2cacc1f595590115a33f5834e)), closes [#1391](https://github.com/kurtosis-tech/kurtosis/issues/1391)
* return the production enclave information if present via GetEnclaves API ([#1395](https://github.com/kurtosis-tech/kurtosis/issues/1395)) ([ef22820](https://github.com/kurtosis-tech/kurtosis/commit/ef22820cad6d98a784bb263435f3dd6e2bbbe31a))


### Bug Fixes

* add scrollbar ([#1400](https://github.com/kurtosis-tech/kurtosis/issues/1400)) ([40aba1d](https://github.com/kurtosis-tech/kurtosis/commit/40aba1ded6ac9b889486c6045332f1bb060ddea8)), closes [#1390](https://github.com/kurtosis-tech/kurtosis/issues/1390)
* bring back args ([#1397](https://github.com/kurtosis-tech/kurtosis/issues/1397)) ([3e1c318](https://github.com/kurtosis-tech/kurtosis/commit/3e1c3188f58a91eb6428ddf17dfa95a0040551c3))
* text off center ([#1407](https://github.com/kurtosis-tech/kurtosis/issues/1407)) ([d845764](https://github.com/kurtosis-tech/kurtosis/commit/d8457640597696d6bfcd6c6e9f12864176bb8b35)), closes [#1406](https://github.com/kurtosis-tech/kurtosis/issues/1406)

## [0.83.9](https://github.com/kurtosis-tech/kurtosis/compare/0.83.8...0.83.9) (2023-09-26)


### Bug Fixes

* rebuild with type bug fix ([#1385](https://github.com/kurtosis-tech/kurtosis/issues/1385)) ([14840b7](https://github.com/kurtosis-tech/kurtosis/commit/14840b73509ddbf6ea4729a20703ea0d77c08da9))
* restart log aggregator on failure ([#1371](https://github.com/kurtosis-tech/kurtosis/issues/1371)) ([7f171ce](https://github.com/kurtosis-tech/kurtosis/commit/7f171ce678ee8915d17c30262930365428d9a4f8))

## [0.83.8](https://github.com/kurtosis-tech/kurtosis/compare/0.83.7...0.83.8) (2023-09-26)


### Bug Fixes

* handle missing arg types ([#1373](https://github.com/kurtosis-tech/kurtosis/issues/1373)) ([5cfea2a](https://github.com/kurtosis-tech/kurtosis/commit/5cfea2a0c62165193d258ad8d5bba48e06d4f5fb))
* Relative import breaks for 'non-main branchs' ([#1364](https://github.com/kurtosis-tech/kurtosis/issues/1364)) ([5496082](https://github.com/kurtosis-tech/kurtosis/commit/549608269f21b2bf886c92263ee60989dc9fb4e1)), closes [#1361](https://github.com/kurtosis-tech/kurtosis/issues/1361)

## [0.83.7](https://github.com/kurtosis-tech/kurtosis/compare/0.83.6...0.83.7) (2023-09-25)


### Features

* improved log experience on the UI. ([#1368](https://github.com/kurtosis-tech/kurtosis/issues/1368)) ([760c7f0](https://github.com/kurtosis-tech/kurtosis/commit/760c7f0a33d3e2e9f777b509562901a8c6f25308))

## [0.83.6](https://github.com/kurtosis-tech/kurtosis/compare/0.83.5...0.83.6) (2023-09-22)


### Features

* implement -n X and -a flags ([#1341](https://github.com/kurtosis-tech/kurtosis/issues/1341)) ([2c6880c](https://github.com/kurtosis-tech/kurtosis/commit/2c6880c9c251843dafacc3a356cc320f5efe85a7))


### Bug Fixes

* enclave manager ui was reading the wrong type fields ([#1367](https://github.com/kurtosis-tech/kurtosis/issues/1367)) ([0bae141](https://github.com/kurtosis-tech/kurtosis/commit/0bae141837324f94841f4e5f311cc7e2bbfa63a1))
* Manually locate docker socket ([#1362](https://github.com/kurtosis-tech/kurtosis/issues/1362)) ([7fe4956](https://github.com/kurtosis-tech/kurtosis/commit/7fe49560b4d99c28e9bda640294bbe0554b57820))

## [0.83.5](https://github.com/kurtosis-tech/kurtosis/compare/0.83.4...0.83.5) (2023-09-21)


### Features

* add service details to EM UI ([#1352](https://github.com/kurtosis-tech/kurtosis/issues/1352)) ([2ccd98d](https://github.com/kurtosis-tech/kurtosis/commit/2ccd98d2066975d7c94c07b6f793878a27c4ed81))
* added ability to lint Starlark packages ([#1360](https://github.com/kurtosis-tech/kurtosis/issues/1360)) ([f4a072c](https://github.com/kurtosis-tech/kurtosis/commit/f4a072cbbdf53614fe752069d12ed8577a6164be)), closes [#1228](https://github.com/kurtosis-tech/kurtosis/issues/1228)
* Support YAML as Package param ([#1350](https://github.com/kurtosis-tech/kurtosis/issues/1350)) ([e33bfe6](https://github.com/kurtosis-tech/kurtosis/commit/e33bfe688e78b15a6468b4d5abf5ad7a5413ca71))


### Bug Fixes

* tail logs from end of log file ([#1339](https://github.com/kurtosis-tech/kurtosis/issues/1339)) ([b8d5816](https://github.com/kurtosis-tech/kurtosis/commit/b8d58169e9c708a71159a87fe52877471c928653))
* warn instead of failing for json log line parse error ([#1336](https://github.com/kurtosis-tech/kurtosis/issues/1336)) ([44b2820](https://github.com/kurtosis-tech/kurtosis/commit/44b282076a6be85e1711ee33cbcc0ae116882ec6))

## [0.83.4](https://github.com/kurtosis-tech/kurtosis/compare/0.83.3...0.83.4) (2023-09-19)


### Features

* Add format flag to kurtosis port print ([#1319](https://github.com/kurtosis-tech/kurtosis/issues/1319)) ([cbbf260](https://github.com/kurtosis-tech/kurtosis/commit/cbbf260d872344c40fd768ed2226510550a8370d))


### Bug Fixes

* scan for first week of existing logs ([#1343](https://github.com/kurtosis-tech/kurtosis/issues/1343)) ([3905782](https://github.com/kurtosis-tech/kurtosis/commit/3905782b45b8b89d77ee70e231a2e04c19ca1bf0))

## [0.83.3](https://github.com/kurtosis-tech/kurtosis/compare/0.83.2...0.83.3) (2023-09-19)


### Features

* CLI service inspect command ([#1323](https://github.com/kurtosis-tech/kurtosis/issues/1323)) ([ec018b9](https://github.com/kurtosis-tech/kurtosis/commit/ec018b94dd276479ae550597a079c79496a6bc4f))


### Bug Fixes

* revert docs ([#1347](https://github.com/kurtosis-tech/kurtosis/issues/1347)) ([efbaf09](https://github.com/kurtosis-tech/kurtosis/commit/efbaf09a86cb3313af03456a8212c92eb8c33120))
* the docs name ([#1345](https://github.com/kurtosis-tech/kurtosis/issues/1345)) ([c3074d0](https://github.com/kurtosis-tech/kurtosis/commit/c3074d06bae6923d3ea2ab7124a2ace1d4a73aad))
* Update testsuite package name to match their location in Github ([#1335](https://github.com/kurtosis-tech/kurtosis/issues/1335)) ([d5218a2](https://github.com/kurtosis-tech/kurtosis/commit/d5218a2e01361d016269607939721cfee08e3a3d))

## [0.83.2](https://github.com/kurtosis-tech/kurtosis/compare/0.83.1...0.83.2) (2023-09-18)


### Features

* disable line numbers and use the name of the file...  ([#1329](https://github.com/kurtosis-tech/kurtosis/issues/1329)) ([1fd0e5a](https://github.com/kurtosis-tech/kurtosis/commit/1fd0e5a10331617e9efdb51e748abac111726bd9))
* Make service start and stop support multiple services ([#1304](https://github.com/kurtosis-tech/kurtosis/issues/1304)) ([1b34b00](https://github.com/kurtosis-tech/kurtosis/commit/1b34b00578b4a989575bbb96ecf9f2562e9db4cf)), closes [#1089](https://github.com/kurtosis-tech/kurtosis/issues/1089)

## [0.83.1](https://github.com/kurtosis-tech/kurtosis/compare/0.83.0...0.83.1) (2023-09-18)


### Features

* changes to the package manager config and the files artifact view ([#1322](https://github.com/kurtosis-tech/kurtosis/issues/1322)) ([e2b0d2b](https://github.com/kurtosis-tech/kurtosis/commit/e2b0d2b50ffa7edd2ff50eeba4c0887aa38ff27b))

## [0.83.0](https://github.com/kurtosis-tech/kurtosis/compare/0.82.24...0.83.0) (2023-09-18)


### ⚠ BREAKING CHANGES

* rename assert to verify ([#1295](https://github.com/kurtosis-tech/kurtosis/issues/1295))
* print a downloaded container images summary after pulling images from remote or locally ([#1315](https://github.com/kurtosis-tech/kurtosis/issues/1315))

### Features

* Clean CLI command now removes unsued Kurtosis images ([#1314](https://github.com/kurtosis-tech/kurtosis/issues/1314)) ([a924f4a](https://github.com/kurtosis-tech/kurtosis/commit/a924f4a7a1b707695bd8ffc7208c1871ea0432ad)), closes [#1131](https://github.com/kurtosis-tech/kurtosis/issues/1131)
* print a downloaded container images summary after pulling images from remote or locally ([#1315](https://github.com/kurtosis-tech/kurtosis/issues/1315)) ([b822870](https://github.com/kurtosis-tech/kurtosis/commit/b822870d10bcb3614ec3cf2fed3db46dd52d9d42)), closes [#1292](https://github.com/kurtosis-tech/kurtosis/issues/1292)


### Code Refactoring

* rename assert to verify ([#1295](https://github.com/kurtosis-tech/kurtosis/issues/1295)) ([651df40](https://github.com/kurtosis-tech/kurtosis/commit/651df406ecf66518005c806d9ccd1bd3260e4af3))

## [0.82.24](https://github.com/kurtosis-tech/kurtosis/compare/0.82.23...0.82.24) (2023-09-14)


### Bug Fixes

* propagate failed img pull error to response line ([#1302](https://github.com/kurtosis-tech/kurtosis/issues/1302)) ([9a4a928](https://github.com/kurtosis-tech/kurtosis/commit/9a4a9284c4dff87dfd861d2bd8878748abe5c3b8))
* revert always pull latest img ([#1306](https://github.com/kurtosis-tech/kurtosis/issues/1306)) ([d4ef19e](https://github.com/kurtosis-tech/kurtosis/commit/d4ef19e1297ae9373263b1392a1a7fead1892af7))

## [0.82.23](https://github.com/kurtosis-tech/kurtosis/compare/0.82.22...0.82.23) (2023-09-14)


### Features

* folks can delete enclaves from the frontend ([#1250](https://github.com/kurtosis-tech/kurtosis/issues/1250)) ([ee11b7c](https://github.com/kurtosis-tech/kurtosis/commit/ee11b7c2a79f153d7d8aa023ee7c03d54065a0c1))
* The current enclave plan is now persisted to the enclave DB every times the execution finishes ([#1280](https://github.com/kurtosis-tech/kurtosis/issues/1280)) ([33d867e](https://github.com/kurtosis-tech/kurtosis/commit/33d867ed62cbf7621aecb775c8f1ba1c01c5d700))


### Bug Fixes

* follow logs ([#1298](https://github.com/kurtosis-tech/kurtosis/issues/1298)) ([9b0bcb7](https://github.com/kurtosis-tech/kurtosis/commit/9b0bcb779bd7c2dd12a359c868f16cf34ec69f13))
* Reset the module global cache on every new interpretation to avoid using outdated modules ([#1291](https://github.com/kurtosis-tech/kurtosis/issues/1291)) ([81c5462](https://github.com/kurtosis-tech/kurtosis/commit/81c54623deb03cdcfb70b075b4a4367e8f4b4e36))
* return after stream err ([#1301](https://github.com/kurtosis-tech/kurtosis/issues/1301)) ([f40559b](https://github.com/kurtosis-tech/kurtosis/commit/f40559b63ca99163336d0ce706d835a8e345e835))

## [0.82.22](https://github.com/kurtosis-tech/kurtosis/compare/0.82.21...0.82.22) (2023-09-11)


### Features

* always pull latest image ([#1267](https://github.com/kurtosis-tech/kurtosis/issues/1267)) ([6706809](https://github.com/kurtosis-tech/kurtosis/commit/670680980957f5eaa5b0ec01ed0ee9b8973d58e7))
* CLI run command option to disable user services port forwarding ([#1252](https://github.com/kurtosis-tech/kurtosis/issues/1252)) ([1c94378](https://github.com/kurtosis-tech/kurtosis/commit/1c94378b9342bbe07647d8c61c47197f5aafcc18)), closes [#1236](https://github.com/kurtosis-tech/kurtosis/issues/1236)
* retain logs for x weeks ([#1235](https://github.com/kurtosis-tech/kurtosis/issues/1235)) ([5f50c8c](https://github.com/kurtosis-tech/kurtosis/commit/5f50c8cc8bf9e5d99570c1c618a5ec367ed194a2))


### Bug Fixes

* inline upgrade warning ([#1254](https://github.com/kurtosis-tech/kurtosis/issues/1254)) ([33ef03a](https://github.com/kurtosis-tech/kurtosis/commit/33ef03a5c3553778d60597cc177893a0c50d6076)), closes [#1244](https://github.com/kurtosis-tech/kurtosis/issues/1244)

## [0.82.21](https://github.com/kurtosis-tech/kurtosis/compare/0.82.20...0.82.21) (2023-09-06)


### Bug Fixes

* the runtime value store now supports `starlark.Bool` value types ([#1249](https://github.com/kurtosis-tech/kurtosis/issues/1249)) ([825f7cd](https://github.com/kurtosis-tech/kurtosis/commit/825f7cdb7b77bfb3a88d658b839141f965ca4fb6))

## [0.82.20](https://github.com/kurtosis-tech/kurtosis/compare/0.82.19...0.82.20) (2023-09-06)


### Bug Fixes

* handle default string value properly ([#1243](https://github.com/kurtosis-tech/kurtosis/issues/1243)) ([6e49059](https://github.com/kurtosis-tech/kurtosis/commit/6e4905973715db54814cf678832a576a89b5fd28))
* Runtime values created by `add_services` were incorrect in the case of a skipped instruction ([#1239](https://github.com/kurtosis-tech/kurtosis/issues/1239)) ([3412486](https://github.com/kurtosis-tech/kurtosis/commit/341248627daa1be920985137080b8705662f1993))

## [0.82.19](https://github.com/kurtosis-tech/kurtosis/compare/0.82.18...0.82.19) (2023-09-05)


### Features

* Add starlark.Value serializer/deserializer for enclave persistence ([#1229](https://github.com/kurtosis-tech/kurtosis/issues/1229)) ([45b9330](https://github.com/kurtosis-tech/kurtosis/commit/45b9330892a6559d75e8859ef6b9b3dff1f09b1b))


### Bug Fixes

* close engine server which is important for triggering the idle enclaves remotion process ([#1219](https://github.com/kurtosis-tech/kurtosis/issues/1219)) ([912e855](https://github.com/kurtosis-tech/kurtosis/commit/912e8551069da797cdbf86e21046f4444ed42b80))
* disabled time.now() ([#1231](https://github.com/kurtosis-tech/kurtosis/issues/1231)) ([26e8d40](https://github.com/kurtosis-tech/kurtosis/commit/26e8d40dc08a9e534af814138eec598f9b21b1ac))
* Does not delete runtime value during idepotent runs ([#1232](https://github.com/kurtosis-tech/kurtosis/issues/1232)) ([a06c247](https://github.com/kurtosis-tech/kurtosis/commit/a06c2473f9f13a3047d09dc74338d27de6ac24f0))
* fix a sneaky segmentation fault where we were propagating a nil error ([#1222](https://github.com/kurtosis-tech/kurtosis/issues/1222)) ([666f4ee](https://github.com/kurtosis-tech/kurtosis/commit/666f4ee677f76f7828c065046c64394322085d74))
* fix a typo in recipe result repository ([#1224](https://github.com/kurtosis-tech/kurtosis/issues/1224)) ([94a4b8b](https://github.com/kurtosis-tech/kurtosis/commit/94a4b8bc5fc79b69845ab4493eb70307cf9d7b0f))

## [0.82.18](https://github.com/kurtosis-tech/kurtosis/compare/0.82.17...0.82.18) (2023-09-01)


### Bug Fixes

* markdown bug ([#1220](https://github.com/kurtosis-tech/kurtosis/issues/1220)) ([2ce4823](https://github.com/kurtosis-tech/kurtosis/commit/2ce4823718033d0c1c61ab1567107b79de039245))

## [0.82.17](https://github.com/kurtosis-tech/kurtosis/compare/0.82.16...0.82.17) (2023-09-01)


### Features

* enable retrieving logs from services in stopped enclaves ([#1213](https://github.com/kurtosis-tech/kurtosis/issues/1213)) ([83c269c](https://github.com/kurtosis-tech/kurtosis/commit/83c269c4a24e377f5446dcda68f0fa4acd4ef7ff))
* Pass enclave name to Starlark global `kurtosis` module ([#1216](https://github.com/kurtosis-tech/kurtosis/issues/1216)) ([c5f2c97](https://github.com/kurtosis-tech/kurtosis/commit/c5f2c97bb349e114e4e7235ce839b1fb9aa00161))
* Persist runtime value store ([#1170](https://github.com/kurtosis-tech/kurtosis/issues/1170)) ([cfec9b3](https://github.com/kurtosis-tech/kurtosis/commit/cfec9b3028d9349cf2b102cb1818cf5e2a41f047))
* track the analytics toggle event ([#1217](https://github.com/kurtosis-tech/kurtosis/issues/1217)) ([10c461f](https://github.com/kurtosis-tech/kurtosis/commit/10c461f7b546cc260540725a64e624d9f99b04f1))

## [0.82.16](https://github.com/kurtosis-tech/kurtosis/compare/0.82.15...0.82.16) (2023-08-31)


### Features

* added description ([#1214](https://github.com/kurtosis-tech/kurtosis/issues/1214)) ([4a95802](https://github.com/kurtosis-tech/kurtosis/commit/4a95802a86c251d01846cc8350f0cf69ca20a412))


### Bug Fixes

* add create enclave name and production mode for enclaves ([#1211](https://github.com/kurtosis-tech/kurtosis/issues/1211)) ([2760f48](https://github.com/kurtosis-tech/kurtosis/commit/2760f486da3941953ef2bfa81bea3115d5a1caa7))
* made forwarding efficient by reducing calls to Kubernetes ([#1200](https://github.com/kurtosis-tech/kurtosis/issues/1200)) ([4df6a1c](https://github.com/kurtosis-tech/kurtosis/commit/4df6a1c2cb12e0dd9e55dbc51d3c6c2d68917ffd))
* Remove 'kurtosis version' depending on the engine ([#1207](https://github.com/kurtosis-tech/kurtosis/issues/1207)) ([ab7dc02](https://github.com/kurtosis-tech/kurtosis/commit/ab7dc027df3949f1479502c2697cc351e3341021)), closes [#1206](https://github.com/kurtosis-tech/kurtosis/issues/1206)

## [0.82.15](https://github.com/kurtosis-tech/kurtosis/compare/0.82.14...0.82.15) (2023-08-30)


### Bug Fixes

* cluster set doesnt get into a weird state of no cluster being set ([#1055](https://github.com/kurtosis-tech/kurtosis/issues/1055)) ([c647035](https://github.com/kurtosis-tech/kurtosis/commit/c6470356e2939d4834d773a085e4b98c1cd44e7f))
* enclave name validation to support valid DNS-1035 label rules ([#1204](https://github.com/kurtosis-tech/kurtosis/issues/1204)) ([74845a8](https://github.com/kurtosis-tech/kurtosis/commit/74845a85e627acc5ffc54162973457869fcc0887))
* make test enclave creation support DNS label rules ([#1202](https://github.com/kurtosis-tech/kurtosis/issues/1202)) ([df61ecc](https://github.com/kurtosis-tech/kurtosis/commit/df61ecc783ade430a434fd129a42c54d1d742263))
* Point to the engine restart command as part of the context switch failure remediation to not conflict with lower level commands ([#1191](https://github.com/kurtosis-tech/kurtosis/issues/1191)) ([f83e513](https://github.com/kurtosis-tech/kurtosis/commit/f83e513a1f2b0e136d4f61e92d4189125f900fd4))
* removed the flaky tests ([#1205](https://github.com/kurtosis-tech/kurtosis/issues/1205)) ([b990674](https://github.com/kurtosis-tech/kurtosis/commit/b990674e20696023c22d1e37fa119bce480ce556))
* this pr fixes the search issue. ([#1201](https://github.com/kurtosis-tech/kurtosis/issues/1201)) ([2a17b1b](https://github.com/kurtosis-tech/kurtosis/commit/2a17b1badd413bee892ff89d2c7697a2c1b32213))

## [0.82.14](https://github.com/kurtosis-tech/kurtosis/compare/0.82.13...0.82.14) (2023-08-29)


### Features

* add creation dialog 2 ([#1194](https://github.com/kurtosis-tech/kurtosis/issues/1194)) ([b586a8a](https://github.com/kurtosis-tech/kurtosis/commit/b586a8a0aa5f84b2f6d5f8bff3079135d4ffde2e))

## [0.82.13](https://github.com/kurtosis-tech/kurtosis/compare/0.82.12...0.82.13) (2023-08-29)


### Bug Fixes

* hyperlane package error  ([#1193](https://github.com/kurtosis-tech/kurtosis/issues/1193)) ([e5946ad](https://github.com/kurtosis-tech/kurtosis/commit/e5946ad50fb3275cd7b26025c4901629427fbc4d))

## [0.82.12](https://github.com/kurtosis-tech/kurtosis/compare/0.82.11...0.82.12) (2023-08-29)


### Bug Fixes

* `react-scripts` dependency changed  on the engine frontend to fix a vulnerability reported by Dependabot ([#1179](https://github.com/kurtosis-tech/kurtosis/issues/1179)) ([e5e15f6](https://github.com/kurtosis-tech/kurtosis/commit/e5e15f6fd90455380d585c7e2cc7ebf434e1ea88))
* handle package catalog edge case ([#1187](https://github.com/kurtosis-tech/kurtosis/issues/1187)) ([2a8a4c8](https://github.com/kurtosis-tech/kurtosis/commit/2a8a4c8a9f902ec3444d4ed1964427b81fc13409))
* ui displays error logs ([#1185](https://github.com/kurtosis-tech/kurtosis/issues/1185)) ([6a2522b](https://github.com/kurtosis-tech/kurtosis/commit/6a2522ba96f9a2ab45dc944f5a1b9bc921d1904d))
* user service streaming logs when the last line is not a completed JSON line ([#1175](https://github.com/kurtosis-tech/kurtosis/issues/1175)) ([fece446](https://github.com/kurtosis-tech/kurtosis/commit/fece446d97f11219595772ffd0b42917676b74e1))

## [0.82.11](https://github.com/kurtosis-tech/kurtosis/compare/0.82.10...0.82.11) (2023-08-29)


### Bug Fixes

* it fixes the log font colour; it shows black now on the cloud em  ([#1182](https://github.com/kurtosis-tech/kurtosis/issues/1182)) ([f13de9f](https://github.com/kurtosis-tech/kurtosis/commit/f13de9f61f1125bccf25535d9e90e7a62ea8375c))

## [0.82.10](https://github.com/kurtosis-tech/kurtosis/compare/0.82.9...0.82.10) (2023-08-28)


### Features

* Production mode enclave ([#1171](https://github.com/kurtosis-tech/kurtosis/issues/1171)) ([84e8c71](https://github.com/kurtosis-tech/kurtosis/commit/84e8c7110731c0237c1a9ec5eb7cacfa22b7337b))

## [0.82.9](https://github.com/kurtosis-tech/kurtosis/compare/0.82.8...0.82.9) (2023-08-28)


### Features

* added package to loader UI ([#1147](https://github.com/kurtosis-tech/kurtosis/issues/1147)) ([9a2edff](https://github.com/kurtosis-tech/kurtosis/commit/9a2edffd095bfa6b0dc760e974aae3ed7939e0c7))

## [0.82.8](https://github.com/kurtosis-tech/kurtosis/compare/0.82.7...0.82.8) (2023-08-28)


### Bug Fixes

* Fix NPE in stacktrace ([#1172](https://github.com/kurtosis-tech/kurtosis/issues/1172)) ([32770ca](https://github.com/kurtosis-tech/kurtosis/commit/32770ca96513d5d4191f6e2f373cadc89120adc9))
* fix the Discord link used by the `dicord` CLI command ([#1177](https://github.com/kurtosis-tech/kurtosis/issues/1177)) ([39d159a](https://github.com/kurtosis-tech/kurtosis/commit/39d159a5141d18a31ddc97d775bccbdd99c2a7ad)), closes [#1174](https://github.com/kurtosis-tech/kurtosis/issues/1174)

## [0.82.7](https://github.com/kurtosis-tech/kurtosis/compare/0.82.6...0.82.7) (2023-08-28)


### Features

* add header to instruct clients not to cache ([#1166](https://github.com/kurtosis-tech/kurtosis/issues/1166)) ([ad27761](https://github.com/kurtosis-tech/kurtosis/commit/ad27761f07306a851526b9424458fe5a54b6cd72))
* Authorize enclave manager requests if the host matches the user's instance host ([#1163](https://github.com/kurtosis-tech/kurtosis/issues/1163)) ([093af33](https://github.com/kurtosis-tech/kurtosis/commit/093af33b4bc9ecf75814ee7c1a2b838379d961fb))

## [0.82.6](https://github.com/kurtosis-tech/kurtosis/compare/0.82.5...0.82.6) (2023-08-26)


### Bug Fixes

* paths ([#1162](https://github.com/kurtosis-tech/kurtosis/issues/1162)) ([e1a9fb0](https://github.com/kurtosis-tech/kurtosis/commit/e1a9fb013acfb8fdc0f638094ab5c596ada0330c))

## [0.82.5](https://github.com/kurtosis-tech/kurtosis/compare/0.82.4...0.82.5) (2023-08-24)


### Bug Fixes

* get service logs tail ([#1156](https://github.com/kurtosis-tech/kurtosis/issues/1156)) ([734b1a8](https://github.com/kurtosis-tech/kurtosis/commit/734b1a8d7431a6e2c35f7abedef803facdffb56d))

## [0.82.4](https://github.com/kurtosis-tech/kurtosis/compare/0.82.3...0.82.4) (2023-08-24)


### Features

* use proxy url ([#1158](https://github.com/kurtosis-tech/kurtosis/issues/1158)) ([7c44373](https://github.com/kurtosis-tech/kurtosis/commit/7c44373fc18ce23117fa7c70155c53a94be09e59))


### Bug Fixes

* Create portal client daemon client only on remote context. ([#1155](https://github.com/kurtosis-tech/kurtosis/issues/1155)) ([b7ae803](https://github.com/kurtosis-tech/kurtosis/commit/b7ae803f24c47046171188b5bba308f5cb0d51f3))

## [0.82.3](https://github.com/kurtosis-tech/kurtosis/compare/0.82.2...0.82.3) (2023-08-24)


### Features

* enable dynamic api host for Enclave Manager and auth enforcement ([#1153](https://github.com/kurtosis-tech/kurtosis/issues/1153)) ([a39706e](https://github.com/kurtosis-tech/kurtosis/commit/a39706e7ab2a7af46afd590b0c9fccf6cd65f4c4))

## [0.82.2](https://github.com/kurtosis-tech/kurtosis/compare/0.82.1...0.82.2) (2023-08-24)


### Features

* make paths relative to support embedding in other contexts ([#1151](https://github.com/kurtosis-tech/kurtosis/issues/1151)) ([74fbe53](https://github.com/kurtosis-tech/kurtosis/commit/74fbe53c07e1dd0c2ae994e2246d1b7033b1bad3))


### Bug Fixes

* lower calls to backend by doing get all services more efficiently ([#1143](https://github.com/kurtosis-tech/kurtosis/issues/1143)) ([a2c0dcc](https://github.com/kurtosis-tech/kurtosis/commit/a2c0dcc0bc3874338ac6fbd5c42bf45163a628dc)), closes [#1074](https://github.com/kurtosis-tech/kurtosis/issues/1074)

## [0.82.1](https://github.com/kurtosis-tech/kurtosis/compare/0.82.0...0.82.1) (2023-08-23)


### Features

* add enclave manager ([#1148](https://github.com/kurtosis-tech/kurtosis/issues/1148)) ([54d94c5](https://github.com/kurtosis-tech/kurtosis/commit/54d94c5e80a2a89d5dbf0e9759871007b9141005))
* Running Kurtosis in Kurtosis cloud doc ([#1142](https://github.com/kurtosis-tech/kurtosis/issues/1142)) ([dbff171](https://github.com/kurtosis-tech/kurtosis/commit/dbff17164c4f37070db2121fcffd04f1866c45a3))


### Bug Fixes

* use connectrpc instead of bufbuild ([#1144](https://github.com/kurtosis-tech/kurtosis/issues/1144)) ([d98bed1](https://github.com/kurtosis-tech/kurtosis/commit/d98bed1b9854624a97c9b2c452071ee476cf282d))

## [0.82.0](https://github.com/kurtosis-tech/kurtosis/compare/0.81.9...0.82.0) (2023-08-22)


### ⚠ BREAKING CHANGES

* Portal remote endpoints and forward port wait until ready option ([#1124](https://github.com/kurtosis-tech/kurtosis/issues/1124))

### Features

* Add possibility to pass env vars to enclave ([#1134](https://github.com/kurtosis-tech/kurtosis/issues/1134)) ([9889e6f](https://github.com/kurtosis-tech/kurtosis/commit/9889e6f126666451d378965e764a80d79ba72443))
* added connect-go for engine ([#879](https://github.com/kurtosis-tech/kurtosis/issues/879)) ([8c0121c](https://github.com/kurtosis-tech/kurtosis/commit/8c0121cac01f53d858bbec90d87bec20f122430d))
* make kurtosis service logs pull from persistent volume ([#1121](https://github.com/kurtosis-tech/kurtosis/issues/1121)) ([8e52a24](https://github.com/kurtosis-tech/kurtosis/commit/8e52a2489cb67707373f802c00dd4f37e7b56931))
* Portal remote endpoints and forward port wait until ready option ([#1124](https://github.com/kurtosis-tech/kurtosis/issues/1124)) ([f4e3e77](https://github.com/kurtosis-tech/kurtosis/commit/f4e3e773463b98f7376ee49b70ab28d9da60caae))
* propogated bridge network ip address ([#1135](https://github.com/kurtosis-tech/kurtosis/issues/1135)) ([04ed723](https://github.com/kurtosis-tech/kurtosis/commit/04ed723c00ac9adb820c56f1db6998eff483f294))


### Bug Fixes

* More flexible context config unmarshaller ([#1132](https://github.com/kurtosis-tech/kurtosis/issues/1132)) ([7892bda](https://github.com/kurtosis-tech/kurtosis/commit/7892bda4fe0a8e0251b9036e9b7cc18843eefcc1))
* Use portal deps from the main branch - Part 2 ([#1138](https://github.com/kurtosis-tech/kurtosis/issues/1138)) ([f0a2353](https://github.com/kurtosis-tech/kurtosis/commit/f0a2353b68a5580552f66fb468a9bd3681b3e7d6))
* Use portal deps from the main branch ([#1136](https://github.com/kurtosis-tech/kurtosis/issues/1136)) ([b9da525](https://github.com/kurtosis-tech/kurtosis/commit/b9da5254edbb291a6a1354c67bcbc357bde71c6c))

## [0.81.9](https://github.com/kurtosis-tech/kurtosis/compare/0.81.8...0.81.9) (2023-08-17)


### Features

* Implements service registration repository in Docker Kurtosis Backend ([#1105](https://github.com/kurtosis-tech/kurtosis/issues/1105)) ([723a14e](https://github.com/kurtosis-tech/kurtosis/commit/723a14e041a74744337aeb38d64fa86306b36883))

## [0.81.8](https://github.com/kurtosis-tech/kurtosis/compare/0.81.7...0.81.8) (2023-08-15)


### Bug Fixes

* make enclave identifier arg passable to service identifier completion provider ([#1107](https://github.com/kurtosis-tech/kurtosis/issues/1107)) ([051bc95](https://github.com/kurtosis-tech/kurtosis/commit/051bc95287fc56ed1af077c8992edeef935bdc57)), closes [#1094](https://github.com/kurtosis-tech/kurtosis/issues/1094)

## [0.81.7](https://github.com/kurtosis-tech/kurtosis/compare/0.81.6...0.81.7) (2023-08-14)


### Features

* add connect-go bindings generation to cloud ([#1090](https://github.com/kurtosis-tech/kurtosis/issues/1090)) ([8ba54d0](https://github.com/kurtosis-tech/kurtosis/commit/8ba54d099e550669d6c3be185880bc1f73ac24f3))


### Bug Fixes

* move where logs aggregator is destroyed ([#1110](https://github.com/kurtosis-tech/kurtosis/issues/1110)) ([aa392f3](https://github.com/kurtosis-tech/kurtosis/commit/aa392f39557afb976a6b74db5c80ffea991b4294))

## [0.81.6](https://github.com/kurtosis-tech/kurtosis/compare/0.81.5...0.81.6) (2023-08-11)


### Features

* add more endpoints to support the Cloud ([#1077](https://github.com/kurtosis-tech/kurtosis/issues/1077)) ([1d70382](https://github.com/kurtosis-tech/kurtosis/commit/1d70382cdefd5361da10c88c64a6c5be81ae3a57))
* enable streaming exec output in container engine (stream exec pt. 1) ([#1043](https://github.com/kurtosis-tech/kurtosis/issues/1043)) ([e8f34ef](https://github.com/kurtosis-tech/kurtosis/commit/e8f34ef3d33cf84499ddb07b461ca87319bef0fc))
* implement new logging architecture v0 ([#1071](https://github.com/kurtosis-tech/kurtosis/issues/1071)) ([c66c148](https://github.com/kurtosis-tech/kurtosis/commit/c66c1480c8f8e6fcc8e17488135ff3d1cb456ffa))
* make enclave namespace and network naming deterministic ([#1100](https://github.com/kurtosis-tech/kurtosis/issues/1100)) ([0d42106](https://github.com/kurtosis-tech/kurtosis/commit/0d42106a015793f7a5d7ede06a54fac58767af7d))
* Persist file artifacts ([#1084](https://github.com/kurtosis-tech/kurtosis/issues/1084)) ([c7b3590](https://github.com/kurtosis-tech/kurtosis/commit/c7b3590a121ef4a9398efe7a5bc479578a04c43f))
* Portal automatic start and stop on context change ([#1086](https://github.com/kurtosis-tech/kurtosis/issues/1086)) ([a6a73d1](https://github.com/kurtosis-tech/kurtosis/commit/a6a73d1c2a03c9d6d9e89b689b86bf170e39f108)), closes [#970](https://github.com/kurtosis-tech/kurtosis/issues/970)
* Update files if already present in enclave ([#1066](https://github.com/kurtosis-tech/kurtosis/issues/1066)) ([1135543](https://github.com/kurtosis-tech/kurtosis/commit/1135543b1dea9ddb2f5419cffd9fd1557e644824))


### Bug Fixes

* Add API key to endpoint ([#1102](https://github.com/kurtosis-tech/kurtosis/issues/1102)) ([64f0c20](https://github.com/kurtosis-tech/kurtosis/commit/64f0c2034405fbaefb7dfb26f63308f055978f53))
* Fix issue with idempotent plan resolution ([#1087](https://github.com/kurtosis-tech/kurtosis/issues/1087)) ([fd48f8f](https://github.com/kurtosis-tech/kurtosis/commit/fd48f8f5f34abe2929b7831ef1453b67eba0b3ca))
* Forward the engine port after verifying that an engine container is running and before initializing the engine client ([#1099](https://github.com/kurtosis-tech/kurtosis/issues/1099)) ([b0b7a3b](https://github.com/kurtosis-tech/kurtosis/commit/b0b7a3b0fa5da07803d1d5b2697ca9805d8147d9))
* update golang docker client to latest ([#1082](https://github.com/kurtosis-tech/kurtosis/issues/1082)) ([724084f](https://github.com/kurtosis-tech/kurtosis/commit/724084f1f0b6d0645990d7b92e41ad6e286f9259))

## [0.81.5](https://github.com/kurtosis-tech/kurtosis/compare/0.81.4...0.81.5) (2023-08-07)


### Features

* Enclave inspect relying on the API container service only ([#1070](https://github.com/kurtosis-tech/kurtosis/issues/1070)) ([da171ea](https://github.com/kurtosis-tech/kurtosis/commit/da171ea6a9350992ec282265ecfa07882dc47c65))


### Bug Fixes

* Fix broken link in docs causing CI build to fail ([#1079](https://github.com/kurtosis-tech/kurtosis/issues/1079)) ([77d8a13](https://github.com/kurtosis-tech/kurtosis/commit/77d8a13e1104eb7b7556a2c3796a2ad5e51f23ec))

## [0.81.4](https://github.com/kurtosis-tech/kurtosis/compare/0.81.3...0.81.4) (2023-08-03)


### Bug Fixes

* Only forward APIC port on remote context. ([#1049](https://github.com/kurtosis-tech/kurtosis/issues/1049)) ([7072b7b](https://github.com/kurtosis-tech/kurtosis/commit/7072b7be2fa0f5a417a3e0ca28c4ca8cb4558a29)), closes [#1039](https://github.com/kurtosis-tech/kurtosis/issues/1039)
* remove historical enclave names from auto complete ([#1059](https://github.com/kurtosis-tech/kurtosis/issues/1059)) ([e63fd88](https://github.com/kurtosis-tech/kurtosis/commit/e63fd88b8bc657f086b631400dd1c70b0f66d1ab))

## [0.81.3](https://github.com/kurtosis-tech/kurtosis/compare/0.81.2...0.81.3) (2023-08-03)


### Bug Fixes

* Pin grpc-file-transfer version on Go SDK ([#1058](https://github.com/kurtosis-tech/kurtosis/issues/1058)) ([36a16ac](https://github.com/kurtosis-tech/kurtosis/commit/36a16ac3b6db9914f3b0a6695535d8ee6ac8ae6b))

## [0.81.2](https://github.com/kurtosis-tech/kurtosis/compare/0.81.1...0.81.2) (2023-08-03)


### Features

* Compute content hash when compressing files artifact ([#1041](https://github.com/kurtosis-tech/kurtosis/issues/1041)) ([510ffe2](https://github.com/kurtosis-tech/kurtosis/commit/510ffe270fea663985b45228e45836fcb575932d))


### Bug Fixes

* Fix comment about sidecar ([#1053](https://github.com/kurtosis-tech/kurtosis/issues/1053)) ([d9b07ea](https://github.com/kurtosis-tech/kurtosis/commit/d9b07ea0a5d609c1191c7e7260a1928ddd1ebd4e))
* Use the local grpc-file-transfer library version ([#1056](https://github.com/kurtosis-tech/kurtosis/issues/1056)) ([59fa980](https://github.com/kurtosis-tech/kurtosis/commit/59fa98013aee05a32a34aa2bef1a153a1a57a29b))

## [0.81.1](https://github.com/kurtosis-tech/kurtosis/compare/0.81.0...0.81.1) (2023-08-02)


### Features

* Print execution steps for kurtosis import ([#1047](https://github.com/kurtosis-tech/kurtosis/issues/1047)) ([44d3b16](https://github.com/kurtosis-tech/kurtosis/commit/44d3b16528a8523f3c20842e373b3b51458fb267))


### Bug Fixes

* Stop local running engine when switching context ([#1040](https://github.com/kurtosis-tech/kurtosis/issues/1040)) ([a8b5606](https://github.com/kurtosis-tech/kurtosis/commit/a8b5606f445cb72126db2bca15efdb294c1d75a0))

## [0.81.0](https://github.com/kurtosis-tech/kurtosis/compare/0.80.24...0.81.0) (2023-08-02)


### ⚠ BREAKING CHANGES

* subnetwork capabilities removed from Kurtosis. Users will have to update their Kurtosis package if they are using subnetwork capabilities ([#1038](https://github.com/kurtosis-tech/kurtosis/issues/1038))

### Features

* subnetwork capabilities removed from Kurtosis. Users will have to update their Kurtosis package if they are using subnetwork capabilities ([#1038](https://github.com/kurtosis-tech/kurtosis/issues/1038)) ([724f713](https://github.com/kurtosis-tech/kurtosis/commit/724f713bd7271dffc10c78dfdc8f5c6c4d42af0d))

## [0.80.24](https://github.com/kurtosis-tech/kurtosis/compare/0.80.23...0.80.24) (2023-08-01)


### Features

* Persistent directories for docker ([#1034](https://github.com/kurtosis-tech/kurtosis/issues/1034)) ([2f909c3](https://github.com/kurtosis-tech/kurtosis/commit/2f909c381c297c75558c9b17ce3974e1d6091b87))
* Persistent directories for Kubernetes ([#1036](https://github.com/kurtosis-tech/kurtosis/issues/1036)) ([4488986](https://github.com/kurtosis-tech/kurtosis/commit/44889866922e728a633573414e5a9ae81310e7c1))


### Bug Fixes

* Remove the temp cert files only after the docker client is initialized ([#1030](https://github.com/kurtosis-tech/kurtosis/issues/1030)) ([1a6bb74](https://github.com/kurtosis-tech/kurtosis/commit/1a6bb747b99bd730cc7c214469d46fff3538fc5f))

## [0.80.23](https://github.com/kurtosis-tech/kurtosis/compare/0.80.22...0.80.23) (2023-07-31)


### Features

* add `cloud add` ([#1015](https://github.com/kurtosis-tech/kurtosis/issues/1015)) ([48aecd0](https://github.com/kurtosis-tech/kurtosis/commit/48aecd05381b9b89fb34da145f9651605ca446d2))


### Bug Fixes

* Fix error swallowing in DefaultServiceNetwork.destroyService ([#987](https://github.com/kurtosis-tech/kurtosis/issues/987)) ([828f366](https://github.com/kurtosis-tech/kurtosis/commit/828f3666d4c0cb27cd83f071204e75143da14348))

## [0.80.22](https://github.com/kurtosis-tech/kurtosis/compare/0.80.21...0.80.22) (2023-07-28)


### Features

* Add starlark converter to kurtosis import ([#1010](https://github.com/kurtosis-tech/kurtosis/issues/1010)) ([8554635](https://github.com/kurtosis-tech/kurtosis/commit/8554635af6990d1b152aa914ef2c595d5f8be802))
* Support resource reservations on Docker compose import ([#1023](https://github.com/kurtosis-tech/kurtosis/issues/1023)) ([e7a5576](https://github.com/kurtosis-tech/kurtosis/commit/e7a5576e1a5dd96b4fdf0b9858caa9394b0572ef))


### Bug Fixes

* truncate output if greater than 64*1024 characters ([#1022](https://github.com/kurtosis-tech/kurtosis/issues/1022)) ([c3e8939](https://github.com/kurtosis-tech/kurtosis/commit/c3e8939811ea4ccafd559cfd9d3705350c6f9fac))

## [0.80.21](https://github.com/kurtosis-tech/kurtosis/compare/0.80.20...0.80.21) (2023-07-28)


### Bug Fixes

* Check if a local engine is running before switching to a remote context and let the user know what to do ([#1011](https://github.com/kurtosis-tech/kurtosis/issues/1011)) ([141247f](https://github.com/kurtosis-tech/kurtosis/commit/141247f46fc5ca11644a35f865c737e96dd3a343))
* fix cpu calculation by getting pre cpu stat ([52a191e](https://github.com/kurtosis-tech/kurtosis/commit/52a191e9e4a1cfaf011ef3b7c0d3d6ea02822756))
* Implement GetEngineLogs in Kubernete backend ([#1005](https://github.com/kurtosis-tech/kurtosis/issues/1005)) ([3d0a3e2](https://github.com/kurtosis-tech/kurtosis/commit/3d0a3e2153da6254f62e53b9f03d9106c57e45a0))
* Log streaming was timing out on docker ([#999](https://github.com/kurtosis-tech/kurtosis/issues/999)) ([d3b6c43](https://github.com/kurtosis-tech/kurtosis/commit/d3b6c434ee3229ba6f433fda5374c0676d690db0))
* make continuity test work ([#1016](https://github.com/kurtosis-tech/kurtosis/issues/1016)) ([c430db2](https://github.com/kurtosis-tech/kurtosis/commit/c430db22616b0684711a79e4326a49102437abe6))
* make resource fetching a parallel operation ([#1012](https://github.com/kurtosis-tech/kurtosis/issues/1012)) ([52a191e](https://github.com/kurtosis-tech/kurtosis/commit/52a191e9e4a1cfaf011ef3b7c0d3d6ea02822756))
* only ask for emails on interactive terminals ([#1018](https://github.com/kurtosis-tech/kurtosis/issues/1018)) ([1bdac73](https://github.com/kurtosis-tech/kurtosis/commit/1bdac73eb07611cb6bcfd987ed4282b7eb06c26e))

## [0.80.20](https://github.com/kurtosis-tech/kurtosis/compare/0.80.19...0.80.20) (2023-07-27)


### Features

* add `kurtosis cloud load to CLI` ([#882](https://github.com/kurtosis-tech/kurtosis/issues/882)) ([b2db8c9](https://github.com/kurtosis-tech/kurtosis/commit/b2db8c98d7b17c96d53c28154739e624fe48a63d))
* ask user for email on first run of Kurtosis ([#1001](https://github.com/kurtosis-tech/kurtosis/issues/1001)) ([0f33b5b](https://github.com/kurtosis-tech/kurtosis/commit/0f33b5b4a3286d9f3a973ad55f7479f17a1782a6))
* Start engine remotely with remote backend config when the context is remote ([#963](https://github.com/kurtosis-tech/kurtosis/issues/963)) ([6816d1f](https://github.com/kurtosis-tech/kurtosis/commit/6816d1f01d99e80609f808b57d2250ebc0b1c8bd))
* validate min cpu & min memory are well under whats available ([#988](https://github.com/kurtosis-tech/kurtosis/issues/988)) ([768e95d](https://github.com/kurtosis-tech/kurtosis/commit/768e95d2dbeb7a554a97cff8b6650e734dccd66a))


### Bug Fixes

* Normalize destroy enclave in all tests ([#976](https://github.com/kurtosis-tech/kurtosis/issues/976)) ([20b635a](https://github.com/kurtosis-tech/kurtosis/commit/20b635a2fc7efc958e7bd7e007b2db65762b8b1c))

## [0.80.19](https://github.com/kurtosis-tech/kurtosis/compare/0.80.18...0.80.19) (2023-07-26)


### Bug Fixes

* Fix docker image pull hanging forever ([#994](https://github.com/kurtosis-tech/kurtosis/issues/994)) ([fd00d79](https://github.com/kurtosis-tech/kurtosis/commit/fd00d79efca2a7d8b3b04ce9d1f4d988dc1d956b))

## [0.80.18](https://github.com/kurtosis-tech/kurtosis/compare/0.80.17...0.80.18) (2023-07-26)


### Features

* Add volume bind support for `kurtosis import` ([#984](https://github.com/kurtosis-tech/kurtosis/issues/984)) ([391c016](https://github.com/kurtosis-tech/kurtosis/commit/391c016ccaa24d454f746179bd096030596bf363))


### Bug Fixes

* CLI args marked as greedy were not greedy ([#975](https://github.com/kurtosis-tech/kurtosis/issues/975)) ([e6ff482](https://github.com/kurtosis-tech/kurtosis/commit/e6ff482cdf6758885ae9a1bdcd3ea6fb5e620a05))

## [0.80.17](https://github.com/kurtosis-tech/kurtosis/compare/0.80.16...0.80.17) (2023-07-26)


### Features

* Add `environment` support for `kurtosis import`  ([#982](https://github.com/kurtosis-tech/kurtosis/issues/982)) ([24e71d1](https://github.com/kurtosis-tech/kurtosis/commit/24e71d1464b9d081056d61f43fde09fba2d8505f)), closes [#981](https://github.com/kurtosis-tech/kurtosis/issues/981)

## [0.80.16](https://github.com/kurtosis-tech/kurtosis/compare/0.80.15...0.80.16) (2023-07-25)


### Features

* folks can now use frontend to view file artifacts and it's content. ([#967](https://github.com/kurtosis-tech/kurtosis/issues/967)) ([fc87c31](https://github.com/kurtosis-tech/kurtosis/commit/fc87c31cd8deecfcba689ef0d0416be017f9ff36))

## [0.80.15](https://github.com/kurtosis-tech/kurtosis/compare/0.80.14...0.80.15) (2023-07-25)


### Features

* Implement V0 of docker import ([#968](https://github.com/kurtosis-tech/kurtosis/issues/968)) ([6f8d90d](https://github.com/kurtosis-tech/kurtosis/commit/6f8d90d526293f676e50243d52aa132ea74447bd))


### Bug Fixes

* Run user service containers in --init mode for Docker ([#965](https://github.com/kurtosis-tech/kurtosis/issues/965)) ([b8989a8](https://github.com/kurtosis-tech/kurtosis/commit/b8989a8112e4f25fed0e595d32a28c45a58a8b1b))

## [0.80.14](https://github.com/kurtosis-tech/kurtosis/compare/0.80.13...0.80.14) (2023-07-24)


### Features

* Add ability to update a running service ([#943](https://github.com/kurtosis-tech/kurtosis/issues/943)) ([42a67f9](https://github.com/kurtosis-tech/kurtosis/commit/42a67f9a3f9d4413f58929867b4e6e61eeeaa25e))
* added create enclave flow ([#962](https://github.com/kurtosis-tech/kurtosis/issues/962)) ([4c931b8](https://github.com/kurtosis-tech/kurtosis/commit/4c931b882e4298cf8d99d88425b0323576f7baf5))
* Idempotent run V1 - services can now be live-updated inside an enclave ([#954](https://github.com/kurtosis-tech/kurtosis/issues/954)) ([a6a118d](https://github.com/kurtosis-tech/kurtosis/commit/a6a118d5b6cc0d3560a5e3abdd8b043397efeced))


### Bug Fixes

* Fix `successfully executed` bug in APIC logs when script fails ([#964](https://github.com/kurtosis-tech/kurtosis/issues/964)) ([32fe63f](https://github.com/kurtosis-tech/kurtosis/commit/32fe63fcb77a8db78b2e1e86be18d3857bfa5fc0))
* no magic string replacement in python packages ([#966](https://github.com/kurtosis-tech/kurtosis/issues/966)) ([8b0fa62](https://github.com/kurtosis-tech/kurtosis/commit/8b0fa623a2c73ec195e2204da5a8463e016e6833))
* the old go download ([#958](https://github.com/kurtosis-tech/kurtosis/issues/958)) ([f1b52ca](https://github.com/kurtosis-tech/kurtosis/commit/f1b52ca98215f090a849e626f934ccd341ad91c3))

## [0.80.13](https://github.com/kurtosis-tech/kurtosis/compare/0.80.12...0.80.13) (2023-07-20)


### Features

* Add autocomplete for file path of artifact files inspect ([#947](https://github.com/kurtosis-tech/kurtosis/issues/947)) ([f72dfce](https://github.com/kurtosis-tech/kurtosis/commit/f72dfce9b755c37dde849f9047ef4a6ca7e59cb2))


### Bug Fixes

* broken symlinks on Kurtosis packages ([#944](https://github.com/kurtosis-tech/kurtosis/issues/944)) ([fbb0aee](https://github.com/kurtosis-tech/kurtosis/commit/fbb0aee6edfce4598b0384aebfe71b1e12b9730c)), closes [#846](https://github.com/kurtosis-tech/kurtosis/issues/846)
* improve frontend ([#940](https://github.com/kurtosis-tech/kurtosis/issues/940)) ([36153e2](https://github.com/kurtosis-tech/kurtosis/commit/36153e2c6e3c332508d6071d2f9101f77cfb6295))
* improved error msg ([#936](https://github.com/kurtosis-tech/kurtosis/issues/936)) ([4f72ae1](https://github.com/kurtosis-tech/kurtosis/commit/4f72ae12409d6ddd8c2e2c6b61770081d9200bde))

## [0.80.12](https://github.com/kurtosis-tech/kurtosis/compare/0.80.11...0.80.12) (2023-07-18)


### Features

* Service count can go up to 1024 in Docker backend ([#919](https://github.com/kurtosis-tech/kurtosis/issues/919)) ([e1dfff1](https://github.com/kurtosis-tech/kurtosis/commit/e1dfff119a0b6635e732e0e09de68b56d6af7d63))

## [0.80.11](https://github.com/kurtosis-tech/kurtosis/compare/0.80.10...0.80.11) (2023-07-18)


### Features

* Add file artifact inspect API do APIC ([#885](https://github.com/kurtosis-tech/kurtosis/issues/885)) ([7ad8155](https://github.com/kurtosis-tech/kurtosis/commit/7ad81553a8056887e1399649536319922a05bdc1))
* Add file inspect command to the CLI ([#905](https://github.com/kurtosis-tech/kurtosis/issues/905)) ([bb36a46](https://github.com/kurtosis-tech/kurtosis/commit/bb36a469925c3a8c00a88c0f5a16995088d26548))
* added run python ([#913](https://github.com/kurtosis-tech/kurtosis/issues/913)) ([365f5cf](https://github.com/kurtosis-tech/kurtosis/commit/365f5cf15399dd0e79f7b82d5ab4ad823def00b5))
* upload files support relative locators ([#930](https://github.com/kurtosis-tech/kurtosis/issues/930)) ([8d60968](https://github.com/kurtosis-tech/kurtosis/commit/8d609686ce78a72f82455592b48eeab94b44c359))


### Bug Fixes

* make service labels more restrictive ([#929](https://github.com/kurtosis-tech/kurtosis/issues/929)) ([a8fb599](https://github.com/kurtosis-tech/kurtosis/commit/a8fb5992d0e60bc50efa8585393048c168e878f0)), closes [#928](https://github.com/kurtosis-tech/kurtosis/issues/928)

## [0.80.10](https://github.com/kurtosis-tech/kurtosis/compare/0.80.9...0.80.10) (2023-07-17)


### Features

* Added enclave pool for improving performance on enclave creation  ([#787](https://github.com/kurtosis-tech/kurtosis/issues/787)) ([d6efa43](https://github.com/kurtosis-tech/kurtosis/commit/d6efa435efeb9989de8f20f1d2d80603b7ef6827))

## [0.80.9](https://github.com/kurtosis-tech/kurtosis/compare/0.80.8...0.80.9) (2023-07-17)


### Features

* added a command that opens the Kurtosis Web UI ([#870](https://github.com/kurtosis-tech/kurtosis/issues/870)) ([5098969](https://github.com/kurtosis-tech/kurtosis/commit/509896934656161002d674fa7c61ccd32c6f899d))
* allow for relative imports from packages ([#891](https://github.com/kurtosis-tech/kurtosis/issues/891)) ([42bedab](https://github.com/kurtosis-tech/kurtosis/commit/42bedab9d45e4988f019dea7ccb2985f058e8199))
* Autocomplete file artifact name on download ([#910](https://github.com/kurtosis-tech/kurtosis/issues/910)) ([2cedd08](https://github.com/kurtosis-tech/kurtosis/commit/2cedd0802a8595c3b299cb844fb42e3495991114))
* Make output directory optional for files download ([#909](https://github.com/kurtosis-tech/kurtosis/issues/909)) ([2543d9a](https://github.com/kurtosis-tech/kurtosis/commit/2543d9ad9c68b86c1c1f09137ca60ddfce785b22))
* Starlark package arguments will be parsed as a deep Struct when `"_kurtosis_parser": "struct"` is passed in the arguments JSON ([#884](https://github.com/kurtosis-tech/kurtosis/issues/884)) ([39ec8c2](https://github.com/kurtosis-tech/kurtosis/commit/39ec8c2d4a867420a76119523eb302dc652adb9b))
* updated golang api sdk to 1.19 ([#908](https://github.com/kurtosis-tech/kurtosis/issues/908)) ([fabbb1c](https://github.com/kurtosis-tech/kurtosis/commit/fabbb1cde6b827ef2255bf184356b2f8a3ba9fbf))


### Bug Fixes

* fixed the log and file artifact issue ([#890](https://github.com/kurtosis-tech/kurtosis/issues/890)) ([7f7fe7b](https://github.com/kurtosis-tech/kurtosis/commit/7f7fe7b2d5dc91ddaa8b088129c5be8de0d9f396))
* pinned go version to 1.19.10 for now ([#907](https://github.com/kurtosis-tech/kurtosis/issues/907)) ([847a37c](https://github.com/kurtosis-tech/kurtosis/commit/847a37c756b50588a567459956f49fcd26d99c28))

## [0.80.8](https://github.com/kurtosis-tech/kurtosis/compare/0.80.7...0.80.8) (2023-07-11)


### Features

* auto assign docs issue to karla ([#834](https://github.com/kurtosis-tech/kurtosis/issues/834)) ([7d0a245](https://github.com/kurtosis-tech/kurtosis/commit/7d0a245fcac4043ab5b780248080b4832b1b0cfe))
* exposing kurtosis frontend v0 ([#833](https://github.com/kurtosis-tech/kurtosis/issues/833)) ([110e910](https://github.com/kurtosis-tech/kurtosis/commit/110e9100ddc69244e7c317ab1e979e15de9f8863))
* Make Run also accept argument other than args dict ([#859](https://github.com/kurtosis-tech/kurtosis/issues/859)) ([9fce411](https://github.com/kurtosis-tech/kurtosis/commit/9fce4112764dfdb135e066e2f54b954f79664b50))


### Bug Fixes

* fixed the output for port print ([#816](https://github.com/kurtosis-tech/kurtosis/issues/816)) ([ede32e7](https://github.com/kurtosis-tech/kurtosis/commit/ede32e795b77387d46ba49e37a6ccc0947fba79a))

## [0.80.7](https://github.com/kurtosis-tech/kurtosis/compare/0.80.6...0.80.7) (2023-07-05)


### Bug Fixes

* Remove existing package directory if it already exists in APIC ([#818](https://github.com/kurtosis-tech/kurtosis/issues/818)) ([4027485](https://github.com/kurtosis-tech/kurtosis/commit/4027485d20917729eb1271387be1317af89ff025))

## [0.80.6](https://github.com/kurtosis-tech/kurtosis/compare/0.80.5...0.80.6) (2023-07-04)


### Features

* Invert USE_INSTRUCTIONS_CACHING feature flag ([#800](https://github.com/kurtosis-tech/kurtosis/issues/800)) ([9a358db](https://github.com/kurtosis-tech/kurtosis/commit/9a358db49d4d222db4c45de62c70e190c6fa7c12))


### Bug Fixes

* fallback to the amd64 image if there's a failure for arm64 image not existing ([#814](https://github.com/kurtosis-tech/kurtosis/issues/814)) ([9cc1033](https://github.com/kurtosis-tech/kurtosis/commit/9cc10332fd67dbe060b883296c7efe5284130b12))

## [0.80.5](https://github.com/kurtosis-tech/kurtosis/compare/0.80.4...0.80.5) (2023-06-30)


### Bug Fixes

* Fix TS proto bindings ([#797](https://github.com/kurtosis-tech/kurtosis/issues/797)) ([7958dba](https://github.com/kurtosis-tech/kurtosis/commit/7958dba5cec3dfb09eb69f24785d33dbd94051d6))
* make dry run return the right return value ([#795](https://github.com/kurtosis-tech/kurtosis/issues/795)) ([be5f6e7](https://github.com/kurtosis-tech/kurtosis/commit/be5f6e75229a3887dc84c7a139aebe84b09fc77d))
* More informative logging for instructions caching ([#785](https://github.com/kurtosis-tech/kurtosis/issues/785)) ([376ac8c](https://github.com/kurtosis-tech/kurtosis/commit/376ac8ceb7085a744c5cf84756b5d2c72a2577f7))

## [0.80.4](https://github.com/kurtosis-tech/kurtosis/compare/0.80.3...0.80.4) (2023-06-28)


### Features

* make the docker network attachable ([#788](https://github.com/kurtosis-tech/kurtosis/issues/788)) ([aeb0b9f](https://github.com/kurtosis-tech/kurtosis/commit/aeb0b9f06749ac42b132f292bc4e24d2b177d472))

## [0.80.3](https://github.com/kurtosis-tech/kurtosis/compare/0.80.2...0.80.3) (2023-06-27)


### Features

* Add minimal support for feature flags in APIC ([#775](https://github.com/kurtosis-tech/kurtosis/issues/775)) ([0858f56](https://github.com/kurtosis-tech/kurtosis/commit/0858f5685365e7d0ab032f362d5ce402c7e5e888))
* added port print functionality in cli for users to quickly check how to access port. ([#778](https://github.com/kurtosis-tech/kurtosis/issues/778)) ([477510b](https://github.com/kurtosis-tech/kurtosis/commit/477510b801a90fce9fcc5bdc403bccc81d505201))
* Implement idempotent run v0 ([#769](https://github.com/kurtosis-tech/kurtosis/issues/769)) ([23b121f](https://github.com/kurtosis-tech/kurtosis/commit/23b121f6ec4e956aa3d1125eeb47bffdb8c136aa))
* Stop and start service support in the CLI ([#767](https://github.com/kurtosis-tech/kurtosis/issues/767)) ([cd4ca05](https://github.com/kurtosis-tech/kurtosis/commit/cd4ca05d17c07892b494b44f23f4c61b1b15d948)), closes [#705](https://github.com/kurtosis-tech/kurtosis/issues/705)

## [0.80.2](https://github.com/kurtosis-tech/kurtosis/compare/0.80.1...0.80.2) (2023-06-26)


### Features

* Add cargo build as part of Kurtosis build ([#774](https://github.com/kurtosis-tech/kurtosis/issues/774)) ([c68fe0a](https://github.com/kurtosis-tech/kurtosis/commit/c68fe0a44c331e72e58762a420fdbc6ec947c9f7))

## [0.80.1](https://github.com/kurtosis-tech/kurtosis/compare/0.80.0...0.80.1) (2023-06-26)


### Features

* Add Rust protobuf bindings ([#765](https://github.com/kurtosis-tech/kurtosis/issues/765)) ([0e47003](https://github.com/kurtosis-tech/kurtosis/commit/0e47003c9f001e31b7a18bc6ea0ddb1d330f0acb))
* added wait to run_sh task. ([#750](https://github.com/kurtosis-tech/kurtosis/issues/750)) ([8c2b697](https://github.com/kurtosis-tech/kurtosis/commit/8c2b697548f06c1f7e8a1474e9ee2cb2922d5dea))
* Implemented rename enclave method in container engine lib ([#755](https://github.com/kurtosis-tech/kurtosis/issues/755)) ([f1570f7](https://github.com/kurtosis-tech/kurtosis/commit/f1570f7e050109c41676e0b9b3aa6b7f251d24ee))
* Persist enclave plan in the Starlark executor memory ([#757](https://github.com/kurtosis-tech/kurtosis/issues/757)) ([2c3d74e](https://github.com/kurtosis-tech/kurtosis/commit/2c3d74e9c88e6b3a980048b6831b23499b4a0a12))
* Start and Stop service Starlark instructions for K8S ([#756](https://github.com/kurtosis-tech/kurtosis/issues/756)) ([fb3e922](https://github.com/kurtosis-tech/kurtosis/commit/fb3e92215fa8062d3a08d1e71ab8572129785688))


### Bug Fixes

* Fix TestStarlarkRemotePackage E2E test to reflect new quickstart ([#773](https://github.com/kurtosis-tech/kurtosis/issues/773)) ([e4dd53f](https://github.com/kurtosis-tech/kurtosis/commit/e4dd53f47ebb6b2efff00b50f035f030169396e5))

## [0.80.0](https://github.com/kurtosis-tech/kurtosis/compare/0.79.0...0.80.0) (2023-06-21)


### ⚠ BREAKING CHANGES

* Applying RFC-1123 standard to service names ([#749](https://github.com/kurtosis-tech/kurtosis/issues/749))

### Features

* Applying RFC-1123 standard to service names ([#749](https://github.com/kurtosis-tech/kurtosis/issues/749)) ([66a5ebe](https://github.com/kurtosis-tech/kurtosis/commit/66a5ebe922559c4d7b8d10b7f7870af6d1700c6b))

## [0.79.0](https://github.com/kurtosis-tech/kurtosis/compare/0.78.5...0.79.0) (2023-06-21)


### ⚠ BREAKING CHANGES

* removed workdir from run_sh and fixed some typos on the doc ([#739](https://github.com/kurtosis-tech/kurtosis/issues/739))

### Features

* allow to pop a shell on Kubernetes ([#748](https://github.com/kurtosis-tech/kurtosis/issues/748)) ([3c706e5](https://github.com/kurtosis-tech/kurtosis/commit/3c706e54f06f60c3950f9c46654ac05b54014dbf))


### Bug Fixes

* removed workdir from run_sh and fixed some typos on the doc ([#739](https://github.com/kurtosis-tech/kurtosis/issues/739)) ([6406f10](https://github.com/kurtosis-tech/kurtosis/commit/6406f10bb1a96cdce429d2c4750977fb86f2d098))
* Support for reconnects in the Gateway port forwarder ([#736](https://github.com/kurtosis-tech/kurtosis/issues/736)) ([4944ccd](https://github.com/kurtosis-tech/kurtosis/commit/4944ccdf32a36786be8816c7e425c08ceccebc9c)), closes [#726](https://github.com/kurtosis-tech/kurtosis/issues/726)
* typos ([#742](https://github.com/kurtosis-tech/kurtosis/issues/742)) ([800e523](https://github.com/kurtosis-tech/kurtosis/commit/800e52364bc62f1dfa1b48bdcae9b01d4d2af7fe))

## [0.78.5](https://github.com/kurtosis-tech/kurtosis/compare/0.78.4...0.78.5) (2023-06-15)


### Features

* added ability for folks to copy files from the one time execution task ([#723](https://github.com/kurtosis-tech/kurtosis/issues/723)) ([f1fcde1](https://github.com/kurtosis-tech/kurtosis/commit/f1fcde148fffe81bc15ea7ab62b00ecd0099e172))
* added run_sh to vscode plugin ([#738](https://github.com/kurtosis-tech/kurtosis/issues/738)) ([337c994](https://github.com/kurtosis-tech/kurtosis/commit/337c9941f6686b2bf7b50416ee7fe71460a8aade))
* Automatically inject the plan object if the first argument of the main function is `plan` ([#716](https://github.com/kurtosis-tech/kurtosis/issues/716)) ([142ce42](https://github.com/kurtosis-tech/kurtosis/commit/142ce42e5a349f468b5ebcbe9ec5f9a752825117))


### Bug Fixes

* Stopping engine not required before switching cluster ([#727](https://github.com/kurtosis-tech/kurtosis/issues/727)) ([af675c1](https://github.com/kurtosis-tech/kurtosis/commit/af675c13a2bcbb10e2619ce513b3c49efa7f642c))

## [0.78.4](https://github.com/kurtosis-tech/kurtosis/compare/0.78.3...0.78.4) (2023-06-13)


### Features

* added run_sh instruction; users can run one time bash task ([#717](https://github.com/kurtosis-tech/kurtosis/issues/717)) ([566144a](https://github.com/kurtosis-tech/kurtosis/commit/566144a5c3cb73f8dc7b8aa13ffb20b9a802edfc))

## [0.78.3](https://github.com/kurtosis-tech/kurtosis/compare/0.78.2...0.78.3) (2023-06-13)


### Features

* Remove `--exec` flag for `kurtosis service shell` ([#712](https://github.com/kurtosis-tech/kurtosis/issues/712)) ([d8bc320](https://github.com/kurtosis-tech/kurtosis/commit/d8bc3206be4ec3d6dec7973c3b31f8746b6089d3))


### Bug Fixes

* add `continue` to avoid segfault on service failing to register ([#719](https://github.com/kurtosis-tech/kurtosis/issues/719)) ([0cebb1f](https://github.com/kurtosis-tech/kurtosis/commit/0cebb1fe22ffd0e0e5e532164c3b0ef658b3ee55))

## [0.78.2](https://github.com/kurtosis-tech/kurtosis/compare/0.78.1...0.78.2) (2023-06-13)


### Bug Fixes

* accept run as the default main function ([#714](https://github.com/kurtosis-tech/kurtosis/issues/714)) ([077cd4c](https://github.com/kurtosis-tech/kurtosis/commit/077cd4c45c7722891d58754fa8b3f4f48c2dfdcb))

## [0.78.1](https://github.com/kurtosis-tech/kurtosis/compare/0.78.0...0.78.1) (2023-06-13)


### Features

* added min/max cpu and memory for kubernetes via starlark ([#689](https://github.com/kurtosis-tech/kurtosis/issues/689)) ([faffc07](https://github.com/kurtosis-tech/kurtosis/commit/faffc071e8617e19bf5b23252a6661cf8b7ff81b))
* use kurtosis service name as the kubernetes service name ([#713](https://github.com/kurtosis-tech/kurtosis/issues/713)) ([b0d6b8e](https://github.com/kurtosis-tech/kurtosis/commit/b0d6b8ebe30f99d1baeaef4d68c08ebd9ca8a9f3))

## [0.78.0](https://github.com/kurtosis-tech/kurtosis/compare/0.77.4...0.78.0) (2023-06-12)


### ⚠ BREAKING CHANGES

* Added `main-file` and `main-function-name` flags to the `kurtosis run` CLI command. These new options were also added in the  `RunStarlarkScript`, `RunStarlarkPackage` and  the `RunStarlarkRemotePackage` SDKs methods, users will have to update the calls. ([#693](https://github.com/kurtosis-tech/kurtosis/issues/693))

### Features

* Added `main-file` and `main-function-name` flags to the `kurtosis run` CLI command. These new options were also added in the  `RunStarlarkScript`, `RunStarlarkPackage` and  the `RunStarlarkRemotePackage` SDKs methods, users will have to update the calls. ([#693](https://github.com/kurtosis-tech/kurtosis/issues/693)) ([1693237](https://github.com/kurtosis-tech/kurtosis/commit/16932374043560daf45689570ec3cbe4e8e174f9))
* random function args ([#703](https://github.com/kurtosis-tech/kurtosis/issues/703)) ([e650a20](https://github.com/kurtosis-tech/kurtosis/commit/e650a20101ee1190b1491ca4ccc8acb58c0bc7dd))
* Start and Stop service Starlark instructions for Docker ([#694](https://github.com/kurtosis-tech/kurtosis/issues/694)) ([10b6b91](https://github.com/kurtosis-tech/kurtosis/commit/10b6b91dc9e8f370bab307297de9b8fe07ca51ce))

## [0.77.4](https://github.com/kurtosis-tech/kurtosis/compare/0.77.3...0.77.4) (2023-06-09)


### Bug Fixes

* make k8s store service files match docker ([#695](https://github.com/kurtosis-tech/kurtosis/issues/695)) ([dc2d8cb](https://github.com/kurtosis-tech/kurtosis/commit/dc2d8cb59d305fa96a2784cd1219352ce235b420))

## [0.77.3](https://github.com/kurtosis-tech/kurtosis/compare/0.77.2...0.77.3) (2023-06-08)


### Features

* Add `kurtosis service exec` command ([#690](https://github.com/kurtosis-tech/kurtosis/issues/690)) ([ece4937](https://github.com/kurtosis-tech/kurtosis/commit/ece49371552f0f9c03fe73b879363c0599f65106))

## [0.77.2](https://github.com/kurtosis-tech/kurtosis/compare/0.77.1...0.77.2) (2023-06-08)


### Features

* added min resource constraint for kubernetes ([#687](https://github.com/kurtosis-tech/kurtosis/issues/687)) ([0aadb91](https://github.com/kurtosis-tech/kurtosis/commit/0aadb912c443c93fe27cebeb727a8ce5f16ced38))
* Label issue based on severity ([#662](https://github.com/kurtosis-tech/kurtosis/issues/662)) ([13b51c6](https://github.com/kurtosis-tech/kurtosis/commit/13b51c6f409432e12b95e9275f5ece788e22989d))


### Bug Fixes

* Auto-restart engine when cluster is updated ([#661](https://github.com/kurtosis-tech/kurtosis/issues/661)) ([479b9f4](https://github.com/kurtosis-tech/kurtosis/commit/479b9f48507def91d76a17731aa84d76c69eff76))
* display service name in exec ([#682](https://github.com/kurtosis-tech/kurtosis/issues/682)) ([6faafea](https://github.com/kurtosis-tech/kurtosis/commit/6faafea86afac1056e529b026743675a5bbfcbf6))
* Fix error propagation in context switch ([#658](https://github.com/kurtosis-tech/kurtosis/issues/658)) ([a7c9bd1](https://github.com/kurtosis-tech/kurtosis/commit/a7c9bd1380d81e7f367daf964021a89086099872))
* Fix typo in the configuration path of the issue labeler workflow ([#667](https://github.com/kurtosis-tech/kurtosis/issues/667)) ([ec6c8e8](https://github.com/kurtosis-tech/kurtosis/commit/ec6c8e885ada06b0adadd44b6d698320a7b43511))
* Fix user service logs when backend is kubernetes ([#678](https://github.com/kurtosis-tech/kurtosis/issues/678)) ([099d046](https://github.com/kurtosis-tech/kurtosis/commit/099d04649f7922adf82dad295f9a701369ee7531))
* fixed the error we see while running the package(s) in dry-mode ([#679](https://github.com/kurtosis-tech/kurtosis/issues/679)) ([af5138c](https://github.com/kurtosis-tech/kurtosis/commit/af5138c1c68ef245c1fd9fce6bd04827ef6c048f))
* Kurtosis shell exec panics if stdin is not terminal ([#686](https://github.com/kurtosis-tech/kurtosis/issues/686)) ([5fad486](https://github.com/kurtosis-tech/kurtosis/commit/5fad4867f9b76498c04c4037ea367161f2c0bb8a))

## [0.77.1](https://github.com/kurtosis-tech/kurtosis/compare/0.77.0...0.77.1) (2023-05-30)


### Features

* Implement PortSpec Wait on Kubernetes backend ([#640](https://github.com/kurtosis-tech/kurtosis/issues/640)) ([7c9989d](https://github.com/kurtosis-tech/kurtosis/commit/7c9989d3119c51c7325de077db1b4f44e4876ce0))
* Run the golang testsuite against K8S (Minikube) ([#653](https://github.com/kurtosis-tech/kurtosis/issues/653)) ([8ddf5ef](https://github.com/kurtosis-tech/kurtosis/commit/8ddf5ef18536b7ae654309f94292ad377373092b))

## [0.77.0](https://github.com/kurtosis-tech/kurtosis/compare/0.76.9...0.77.0) (2023-05-25)


### ⚠ BREAKING CHANGES

* Add Kubernetes implementation ([#638](https://github.com/kurtosis-tech/kurtosis/issues/638))

### Features

* Add Kubernetes implementation ([#638](https://github.com/kurtosis-tech/kurtosis/issues/638)) ([8ad708b](https://github.com/kurtosis-tech/kurtosis/commit/8ad708bca139c79312de60643db1691938f55861))

## [0.76.9](https://github.com/kurtosis-tech/kurtosis/compare/0.76.8...0.76.9) (2023-05-23)


### Bug Fixes

* 'engine stop' now waits for engine to report STOPPED status ([#635](https://github.com/kurtosis-tech/kurtosis/issues/635)) ([e16e123](https://github.com/kurtosis-tech/kurtosis/commit/e16e12304a260c0b6bcbcb6ab119e5b8380880db))

## [0.76.8](https://github.com/kurtosis-tech/kurtosis/compare/0.76.7...0.76.8) (2023-05-23)


### Features

* Return error on SDK if Starlark run on any step  ([#634](https://github.com/kurtosis-tech/kurtosis/issues/634)) ([8a01cff](https://github.com/kurtosis-tech/kurtosis/commit/8a01cfffc92c47d44d0a73593bf91d4c990f72ed))


### Bug Fixes

* Make printWarningIfArgumentIsDeprecated unit test deterministic ([#633](https://github.com/kurtosis-tech/kurtosis/issues/633)) ([46bbee5](https://github.com/kurtosis-tech/kurtosis/commit/46bbee5dcd67346f0007d6d83326fd9200fa9dda))
* Rollback to previous cluster when cluster set fails ([#631](https://github.com/kurtosis-tech/kurtosis/issues/631)) ([0e212c9](https://github.com/kurtosis-tech/kurtosis/commit/0e212c93f05fc174a6ad47bafb25975e0b95b892))

## [0.76.7](https://github.com/kurtosis-tech/kurtosis/compare/0.76.6...0.76.7) (2023-05-17)


### Bug Fixes

* Exclude resources dir from the internal testsuites ([#622](https://github.com/kurtosis-tech/kurtosis/issues/622)) ([ffd2031](https://github.com/kurtosis-tech/kurtosis/commit/ffd203174db8d515752ddf832a8dbfc924687520))
* Remove the GRPC proxy port from the engine and from the APIC ([#626](https://github.com/kurtosis-tech/kurtosis/issues/626)) ([de284be](https://github.com/kurtosis-tech/kurtosis/commit/de284bed4f9031e51fb4ccafc934e39bea3879d5))
* set MTU to 1440 to fix GitPod networking ([#627](https://github.com/kurtosis-tech/kurtosis/issues/627)) ([19ec18e](https://github.com/kurtosis-tech/kurtosis/commit/19ec18e4174555b51c917e13f34f7275c6ddab1a))

## [0.76.6](https://github.com/kurtosis-tech/kurtosis/compare/0.76.5...0.76.6) (2023-05-12)


### Bug Fixes

* ips are on the range 172.16.0.0/16 ([#618](https://github.com/kurtosis-tech/kurtosis/issues/618)) ([b48cb73](https://github.com/kurtosis-tech/kurtosis/commit/b48cb73dadffdb23922c73b68fed1485840eb846))

## [0.76.5](https://github.com/kurtosis-tech/kurtosis/compare/0.76.4...0.76.5) (2023-05-11)


### Features

* Support path argument autocomplete in all CLI commands ([#607](https://github.com/kurtosis-tech/kurtosis/issues/607)) ([e5a5fe1](https://github.com/kurtosis-tech/kurtosis/commit/e5a5fe1f4c690a4ceeea63e718fb4c446e921940))

## [0.76.4](https://github.com/kurtosis-tech/kurtosis/compare/0.76.3...0.76.4) (2023-05-11)


### Features

* Add Windows support for CLI ([#608](https://github.com/kurtosis-tech/kurtosis/issues/608)) ([4cc1c56](https://github.com/kurtosis-tech/kurtosis/commit/4cc1c56e3cebf41c5a033df718938a4d805a3400))
* added sign-up for kcloud ([#591](https://github.com/kurtosis-tech/kurtosis/issues/591)) ([16641e9](https://github.com/kurtosis-tech/kurtosis/commit/16641e9ed0947ea34d44b0c521b429ace5ab5b50))
* Help developers to work across the project modules ([#596](https://github.com/kurtosis-tech/kurtosis/issues/596)) ([e7f845e](https://github.com/kurtosis-tech/kurtosis/commit/e7f845ecd67c8218b28ff284b12ac18949108364))
* return deprecation warnings to users in yellow in colour. ([#586](https://github.com/kurtosis-tech/kurtosis/issues/586)) ([7609fd8](https://github.com/kurtosis-tech/kurtosis/commit/7609fd8c77994875eae77fd458f1744f267c17fb))


### Bug Fixes

* Enable autocomplete for the `files upload` path argument ([#598](https://github.com/kurtosis-tech/kurtosis/issues/598)) ([be52f9e](https://github.com/kurtosis-tech/kurtosis/commit/be52f9e73c5cd63e09f5c2343add165886bd7313))
* kurtosis --&gt; kurtosistech in readme ([#604](https://github.com/kurtosis-tech/kurtosis/issues/604)) ([d6c2ea2](https://github.com/kurtosis-tech/kurtosis/commit/d6c2ea2f6f8127c799701707e65c7697c8354452))
* Pipe metric reporting logs to logger instead of stderr ([#576](https://github.com/kurtosis-tech/kurtosis/issues/576)) ([7060473](https://github.com/kurtosis-tech/kurtosis/commit/7060473563f12b9d097aeb20eb3e4c5cf3e58d55))
* Refresh the README dev instructions ([#595](https://github.com/kurtosis-tech/kurtosis/issues/595)) ([0c71fac](https://github.com/kurtosis-tech/kurtosis/commit/0c71fac3ae3a36fdf6df56e567b3ba184a6756b6))
* rename cloud--&gt;kloud in readme ([#602](https://github.com/kurtosis-tech/kurtosis/issues/602)) ([a998d39](https://github.com/kurtosis-tech/kurtosis/commit/a998d39a3511cf6ba84759f4b91cb20795cefd3d))
* Support redirects with cookies in the user support URLs validation test ([#600](https://github.com/kurtosis-tech/kurtosis/issues/600)) ([ce9718e](https://github.com/kurtosis-tech/kurtosis/commit/ce9718ed55e60cd227f036149da0c410ba99be09)), closes [#599](https://github.com/kurtosis-tech/kurtosis/issues/599)

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

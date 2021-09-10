# TBD

# 0.2.3
### Features
* Verify that, when starting a container with `shouldPublishAllPorts` == `true`, each used port gets exactly one host machine port binding

### Fixes
* Fixed ports not getting bound when running a container with an `EXPOSE` directive in the image Dockerfile
* Fixed broken CI

# 0.2.2
### Fixes
* Fixed bug in markdown-link-check property

# 0.2.1
### Features
* Set up CircleCI checks

# 0.2.0
### Breaking Changes
* Added an `alias` field to `CreateAndStartContainer`

# 0.1.0
* Init commit

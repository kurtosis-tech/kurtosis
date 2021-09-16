# TBD
### Features
* Made `PullImage` a public function on `DockerManager`

# 0.2.7
### Features
* Added extra trace logging to Docker manager

# 0.2.6
### Changes
* Added extra debug logging to hunt down an issue where the `defer` to remove a container that doesn't start properly would try to remove a container with an empty container ID (which would return a "not found" from the Docker engine)

# 0.2.5
### Fixes
* Fixed a bug where not specifying an image tag (which should default to `latest`) wouldn't actually pull the image if it didn't exist locally

# 0.2.4
### Fixes
* Add extra error-checking to handle a very weird case we just saw where container creation succeeds but no container ID is allocated

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

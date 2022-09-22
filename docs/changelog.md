# TBD
### Breaking Changes
* Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
    * Users will need to update all the `CreateEngine` calls removing this new parameter

### Features
* Created the `LogsDatabase` object in `container-engine-lib`
* Created the `LogsCollector` object in `container-engine-lib`

### Changes
* Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`

# 0.49.1
### Fixes
* Attempting to fix the release version
### Changes
* Added container-engine-lib

# 0.49.0

### Changes
* This version is a dummy version to set the minimum. We pick a version greater than the current version of the CLI (0.29.1). 

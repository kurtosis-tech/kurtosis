# TBD

### Fixes
- Hashed the `package_id`
- Hashed the `package_id` in the enclave size tracking event
- Add `is_subnetworking_enabled` to CreateEnclaveEvent

# 0.3.0

### Changes
- Removed metrics for tracking module related events

### Breaking Changes
- Removed metrics for tracking module related events
  - Clients using module tracking should stop using it

### Features
- Added `is_ci` with every event that gets published
- Added `os` and `arch` with every event that gets published
- Added the `backend` to every call
- Added a `kurtosis-run-finished` event to track number of services created by a package when the package is finished running

# 0.2.2

### Features
- Added metrics tracking for `RunStarlarkScript` and `RunStarlarkPackage`

# 0.2.1
### Features
* Capture the raw container image that module load occurs with

### Changes
* Don't error when things we expect to be empty are nil, so that we still at least get the data

### Fixes
* Fix bug where a module container image without `:` would throw an error which meant the event wouldn't be tracked
* Fixed a bug where the Segment client's `ExecuteModule` event would close the underlying client

# 0.2.0
### Changes
* Moved `SegmentClient` and do `DoNothingClient` from their own package to client's package

### Breaking Changes
* Added new `callback` argument in `CreateMetricsClient` method
  * Users should send a callback object that implements the `analytics.Callback` interface
* `CreateMetricsClient` returns a client's close function that should be used to close the client
  * Users should receive the function and them can execute it in the next line in the code using a 'defer' sentence
* Metrics client `Close` function is not exported anymore, now is a private method
  * Users should use the function returned byt the `CreateMetricsClient` to close the client 

# 0.1.2
### Changes
* Updated client's queue flush interval from 10 minutes to 5 second

# 0.1.1
### Features
* Added `MetricsClient` interface to define Kurtosis metrics abstraction behaviour
* Added `SegmentClient` implementation of the `MetricsClient` using Segment provider
* Added `DoNothingClient` implementation of the `MetricsClient` used when users decide reject to send metrics
* Added `Event` object to set the fields involve in a Kurtosis Event. `Category` and `Action` fields are mandatory
* Added event types to centralize Kurtosis events data
* Added `Source` type to define the metrics application source
* Added metrics client creator func to create default metrics type depending on the passed arguments

# 0.1.0
* Initial commit

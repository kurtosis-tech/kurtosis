# TBD
* removing freeHostPortTracker and all host-container port mappings

# 0.3.1
* explicitly specify service IDs in network configurations

# 0.3.0
* Stop the node network after the test controller runs
* Rename ServiceFactory(,Config) -> ServiceInitializer(,Core)
* Fix bug with not actually catching panics when running a test
* Fix a bug with the TestController not appropriately catching panics
* Log test result as soon as the test is finished
* Add some extra unit tests
* Implement controller log propagation
* Allow services to declare arbitrary file-based dependencies (necessary for staking)

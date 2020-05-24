* [ ] TestSuiteRunner needs to throw an error if a user double-registers a test with the same name
* [x] Implement FreeHostPortTracker
    1. [x] Create thread-safe object that tracks free ports with:
        * [x] Method to dole out free ports
        * [x] Method to release ports
    2. [x] Create in the TestSuiteRunner's RunTests method
    3. [x] Pass to JsonRpcServiceNetworkConfig (probably during CreateAndRun)
* [ ] Make JsonRpcServiceNetworkConfig's CreateAndRun method create a Docker network with a UUID
    * [x] Use a UUID to get a unique network name
    * [ ] Use Docker client to actually create a network
* [x] Create a multi-node Ava network config provider
    * [x] Modify GeckoServiceConfig to allow for waiting for boot nodes to start up
        * [x] Implement inter-node dependencies for JsonRpcServiceConfig
            1. [x] Implement a fluent Builder for JsonRpcServiceConfig
                1. [x] Switch current constructor to Builder
                2. [x] Implement a method to add nodes
            2. [x] Modify the Builder's AddService method to also allow depending on existing service
            3. [x] Modify CreateAndRun command to start nodes in the correct order, passing the LivenessRequests from dependencies to the dependents
        * [x] Make GeckoServiceConfig modify its start command based on if it had boot nodes passed in or not
            * [NA] Inside the GetStartCommand method, wrap the actual start command with an image-specific busy loop that checks to make sure at least one boot node is up (NOTE: we should really just have them change Gecko to keep retrying boot nodes until it gets one!!!)
* [ ] Use nat.Port objects to represent ports
    * [ ] Switch JsonRpcServiceNetwork's port fields
    * [ ] Switch JsonRpcServiceConfig's port field
* [ ] Ascertain whether we should be passing structs around by value, or by reference

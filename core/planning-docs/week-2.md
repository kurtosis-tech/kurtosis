* [ ] Fix JsonRpcServiceNetworkConfig to actually fill in the other values of JsonRpcServiceNetwork
* [x] Implement extra Ava-specific params on GeckoServiceConfig:
    * [x] Snow quorum
    * [x] Snow sample
* [ ] Rename GeckoServiceConfig to reflect its a) Docker image base (which affects its start command) and b) API version????
* [ ] Upgrade code to pull images from Docker Hub, not just locally
* [ ] Implement graceful cleanup of Docker containers in TestSuiteRunner that's impossible to skip (even if an exception is thrown)

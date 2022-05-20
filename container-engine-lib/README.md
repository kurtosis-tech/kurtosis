Container Engine Lib
====================
This library abstracts away interactions with the container engine (be it Docker or Kubernetes) via a `KurtosisBackend` interface. Users should call `GetLocalDockerKurtosisBackend` (or the Kubernetes equivalent when it exists).

TODO documentation on:
* What KurtosisBackend's purpose is
* The CRUD methods, and how they're necessarily backend-agnostic
* Kurtosis objects can be composed of multiple resources
* We use the `XXXXX(Kubernetes|Docker)Resources` to represent the various ones
* There are "canonical" Kubernetes/Docker resources that define a given Kurtosis object
    * Created first, destroyed last
* Processing pipeline for CRUD:
    1. Run `getMatchingXXXXXXX(Docker|Kubernetes)Resources` to gather tentative list of resources from Kubernetes/Docker as `XXXXX(Kubernetes|Docker)Resources`
    1. Run extraction commant, to get the Kurtosis object out of the Docker/Kubernetes resources
    1. Apply the filters to the Kurtosis objects
    1. (do function-specific logic, e.g. stopping, updating, etc.)
* Processing pipeline is necessary because the RUD methods all take in filters on Kurtosis objects, so we have to parse the Kubernetes/Docker objects to Kurtosis objects before we can apply the filter

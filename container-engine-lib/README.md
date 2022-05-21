Container Engine Lib
====================
This library abstracts away interactions with the container engine (be it Docker or Kubernetes) via a `KurtosisBackend` interface. Users should call `GetLocalDockerKurtosisBackend` (or the Kubernetes equivalent when it exists).

**NOTE:** To test locally, you can use the `main.go` file which is set up for plugging in whatever logic you want to test!

TODO documentation on:
* How you can use `main.go` to debug locally!!
* What KurtosisBackend's purpose is
* The CRUD methods, and how they're necessarily backend-agnostic
* Kurtosis objects can be composed of multiple resources
* Explanation of prefiltering on Kubernetes resources (get rid of it)
* Explanation of the XXXXXFilters:
    * They're DISJUNCTIVE within a filter layer, but CONJUNCTIVE across the layers
    * This actually makes the filters simpler, because it makes the filters purely subtractive: each XXXXFilters can only reduce the set of matching objects, and each layer within the XXXXFilters only reduces the set of matching objects
* We use the `XXXXX(Kubernetes|Docker)Resources` to represent the various ones
* There are "canonical" Kubernetes/Docker resources that define a given Kurtosis object
    * Created first, destroyed last
* Processing pipeline for CRUD:
    1. Run `getMatchingXXXXXXX(Docker|Kubernetes)Resources` to gather tentative list of resources from Kubernetes/Docker as `XXXXX(Kubernetes|Docker)Resources`
    1. Run extraction commant, to get the Kurtosis object out of the Docker/Kubernetes resources
    1. Apply the filters to the Kurtosis objects
    1. (do function-specific logic, e.g. stopping, updating, etc.)
* Processing pipeline is necessary because the RUD methods all take in filters on Kurtosis objects, so we have to parse the Kubernetes/Docker objects to Kurtosis objects before we can apply the filter

The 4 patterns:
1. impleenting DockerKurtosisBackend: there's no pattern at all; everyone just writes whatever helper functions they feel they need
2. midway through, we discover that if you hide things behind getMatching(filters) it makes your life much easier, but getMatching returns `map[docker-resource-id-as-string]*KurtosisObject`
3. we hit Kubernetes, and now there are multiple resources per Kurtosis object, and things explode a bit again
4. We discover The Pipeline of grouping stuff together, except it returns `map[xxxxxGUID]*KurtosisObject, map[xxxxGUID]xxxxxKubernetesResources`
5. It turns out that dealing with the two maps is a pain (have to do cross dereferences, etc.), so we group them together into a `xxxxxxObjectsAnd(Kubernetes|Docker)Resources`

Container Engine Lib
====================
This library abstracts away interactions with the container engine (be it Docker or Kubernetes) via a `KurtosisBackend` interface. Users should call `GetLocalDockerKurtosisBackend` (or the Kubernetes equivalent when it exists).

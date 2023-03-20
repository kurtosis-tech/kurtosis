Contexts config store
=====================

The contexts config store module is responsible for storing information about the different contexts each
Kurtosis installation is tracking. 

Right now, the contexts config object holds the following information:
1. The configuration for every context that has been configured. A context can be:
   1. a local-only context, no remote Kurtosis server. In this case it contains very little configuration
   1. a remote-backend context, in which case it contains all the information needed to connect to the remote Kurtosis 
server
1. The UUID of the context that is currently selected

The responsibilities of the context config store are to:
1. Load the current contexts config
1. Write to the current contexts config the following operations:
   1. Add a new context to the list of currently configured contexts
   1. Remove a context currently configured
   1. Switch the currently selected context to another context present in the list of configured context

The contexts config is right now persisted to disk, written to a file. We use XCD to manage file path independently of 
the platform on which Kurtosis is used.

:warning: It provides a certain level of thread safety but _within the same process_. I.e. if you have different 
processes both using the store as a library, thread safety will not be guaranteed. Right now it doesn't seem to be 
necessary, but if we ever need this it can be implemented using a `.lock` file on disk.

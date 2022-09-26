Startosis Documentation
=======================

[WIP documentation] 

Startosis is Kurtosis programming language to build an enclave from scratch. It is based on Starlark and provides a 
unique set of Kurtosis bindings to interact with an enclave.

Startosis Reference
-------------------

#### add_service(serviceId services.ServiceID, containerConfig services.ContainerConfig)
Adds a service to the enclave with the specified serviceId and container config.

Startosis Execution layer
-------------------------

Here is a high level explanation of how Startosis is compiled and executed inside Kurtosis engine. See the glossary
below for the definition of each technical term.

The Startosis engine is a 2 process step. First it compiles the Startosis script, second it executes it.

#### Compiler

The compiler is responsible for taking a [Startosis script](#Startosis script) as input and transforming it into a
series of [Kurtosis instructions](#Kurtosis instruction).

The way it does that is by interpreting the Startosis script using a simple 
[starlark-go](https://github.com/google/starlark-go) interpret with custom bindings. The bindings will populate a queue
of [Kurtosis instructions](#Kurtosis instruction).

For now, no validation is run on this list of instructions. This is something that will be implemented in a future
version.

#### Executor

The executor is responsible for taking a set of valid Kurtosis instructions and execute them one after the other against
the Kurtosis backend.

The way is works is very easy. Each instruction produced by the compiler implements an interface with an `Execute()`.
The Executor goes through the queue of instructions, execute them one by one. Each instruction returns an exit code (0
or 1). The executor stops when an instruction fails or when it reaches the end of the queue, whichever comes first.

```
                  Startosis script
                         |
                         V
              /------------------------\
              |       COMPILER         |
              \------------------------/
                         |
                         V
            Queue of Kurtosis instructions
                         |
                         V
              /------------------------\
              |       EXECUTOR         |
              \------------------------/
                         |
                         V
         Queue of Kurtosis instruction outputs
```



### Glossary

#### Startosis script
A Startosis script is a subclass of a Starlark script using Kurtosis bindings.

For example, the following is a valid Startosis script:
```starlark
def addSingleExampleDatastoreServer(serviceNumber):
    serviceId = image + "/" + serviceNumber
    print("Adding service " + serviceId)
    add_service(serviceId, imageId) # Startosis specific function
    
def addMultipleExampleDatastoreServer(numberOfServicesToAdd):
    for i in range(numberOfServicesToAdd):
        addExampleDatastoreServer(i)

imageId = kurtosistech/example-datastore-server
print("Adding 3 nodes for service: " + image)
addMultipleExampleDatastoreServer(3)
```

#### Kurtosis instruction
A Kurtosis instruction is a *unique*, *self-sufficient* directive that can be executed by the Kurtosis engine.

For example, the following is a valid *set* of Kurtosis instruction*s*:
```starlark
print("Adding 3 nodes for service: kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/1", "kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/2", "kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/3", "kurtosistech/example-datastore-server")
```
As an instruction is unique and self-sufficient, it has no context and therefore use of variables and functions is
prohibited.

#### Kurtosis instruction output
A Kurtosis instruction can produce an output. For now, the output is simple and corresponds to the status exit code of
the instruction: 0 for success, 1 for failure.

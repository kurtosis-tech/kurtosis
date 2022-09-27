Startosis Documentation
=======================

[WIP documentation] 

Startosis is Kurtosis programming language to build an enclave from scratch. It is based on Starlark and provides a 
unique set of Kurtosis bindings to interact with an enclave.

Startosis Reference
-------------------

#### `add_service(service_id: ServiceID, service_config: ServiceConfig)`
Adds a service to the enclave with the specified ID and config.

Startosis Execution layer
-------------------------

Here is a high level explanation of how Startosis is interpreted and executed inside Kurtosis engine. See the glossary
below for the definition of each technical term.

The Startosis engine is a 2 process step. First it interprets the Startosis script, second it executes it.

#### Interpreter

The interpreter is responsible for taking a [Startosis script][startosis_script] as input and transforming it into a
series of [Kurtosis instructions][kurtosis_instruction].

The way it does that is by interpreting the Startosis script using 
[starlark-go][starlark_go_home] with custom Kurtosis bindings. To learn more about starlark builtin
features, see the [language reference page][starlark_go_ref].

The interpretation can fail if the Startosis script is invalid (syntax errors). In this case, it will return an 
`InterpretationError` to the user and the engine will exit.

When the interpretation is successful it returns a queue of executable [Kurtosis instructions][kurtosis_instruction].

For now, no validation is run on this list of instructions. This is something that will be implemented in a future
version.

#### Executor

The executor is responsible for taking the list of valid Kurtosis instructions and execute them one after the other against
the Kurtosis backend.

The way it works is very easy. Each instruction produced by the interpreter implements an interface with an `Execute()`
function. The Executor goes through the queue of instructions, executing them one by one. If an instruction fails, the executor
exits immediately and returns an `ExecutionError` to the user. 
When all instructions are successful, the engine returns.

```
                  Startosis script
                         |
                         V
              /------------------------\
              |       INTERPRETER      |
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
def add_single_example_datastore_server(service_number):
    service_id = image + "/" + service_umber
    print("Adding service " + service_idd)
    add_service(service_id, image_id) # Startosis specific function
    
def add_multiple_example_datastore_server(number_of_services_to_add):
    for i in range(number_of_services_to_add):
        add_single_example_datastore_server(i)

image_id = "kurtosistech/example-datastore-server"
print("Adding 3 nodes for service: " + image)
add_multiple_example_datastore_server(3)
```

#### Kurtosis instruction
A Kurtosis instruction is a *unique*, [*context-free*][context_free_grammar] directive that can be executed by the
Kurtosis engine.

For example, the following is a valid *series* of Kurtosis instruction*s*:
```starlark
print("Adding 3 nodes for service: kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/1", "kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/2", "kurtosistech/example-datastore-server")
add_service("kurtosistech/example-datastore-server/3", "kurtosistech/example-datastore-server")
```
As an instruction is unique and self-sufficient, it has no context and therefore use of variables and functions is
prohibited.


<!-- ONLY LINKS BELOW HERE -->
[context_free_grammar]: https://en.wikipedia.org/wiki/Context-free_grammar
[kurtosis_instruction]: #kurtosis-instruction
[starlark_go_home]: https://github.com/google/starlark-go
[starlark_go_ref]: https://github.com/google/starlark-go/blob/master/doc/spec.md#expressions
[startosis_script]: #startosis-script


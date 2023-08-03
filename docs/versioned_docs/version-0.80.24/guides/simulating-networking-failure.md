---
title: Simulating A Networking Failure
sidebar_label: Simulating a networking failure
---

Using Kurtosis you can control which services can communicate inside and across [subnetworks][subnetworks-reference]. Kurtosis provides arbitrarily fine grained control over which services can communicate with each other. This can be useful to simulate network faults, [partition tolerance][cap-theorem] or even partial network outages, including package drops.

This guide illustrates the concept of Kurtosis subnetworks by simluating a partial networking failure within a distributed system.

Step 1: Start all services in an enclave
-----------------------------------------

The example system is composed of a Cassandra cluster with 3 nodes, 2 redundant services serving the main application and a single port proxy to load balance the requests.

```python
MAIN_SUBNETWORK = "main_subnetwork"
SECONDARY_SUBNETWORK = "secondary_subnetwork"

def run(args):
    add_service(
        name="cassandra_node_1",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
    add_service(
        name="cassandra_node_2",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
    add_service(
        name="cassandra_node_3",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
    add_service(
        name="app_service_1",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
    add_service(
        name="app_service_2",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
    add_service(
        name="single_port_proxy",
        config=ServiceConfig(
            ...
            subnetwork=MAIN_SUBNETWORK,
        ),
    )
```

When all the services are added to the Kurtosis enclave, it will look something like:

```
                     Kurtosis enclave                  
  -----------------------------------------------------
 |                   main_subnetwork                   |
 |-----------------------------------------------------|
 |                                                     |
 |                  cassandra_node_1                   |
 |                  cassandra_node_2                   |
 |                  cassandra_node_3                   |
 |                    app_service_1                    |
 |                    app_service_2                    |
 |                  single_port_proxy                  |
 |                                                     |
  -----------------------------------------------------
```

At the start, all connections are allowed.
- all the Cassandra nodes can communicate with each other
- the two services can reach all Cassandra nodes
- the single port proxy is playing its role of dispatching requests to both `app_service_1` and `app_service_2`

Setp 2: Add a second subnetwork by updating services
----------------------------------------------------

Subnetworks make it easy to simulate a network failure that completely isolates `cassandra_node_2` and `app_service_2` from the rest of the system.

First,`cassandra_node_2` and `app_service_2` need to be assigned to a distinct subnetwork, `secondary_subnetwork`. This can be done by updating those services.

This can be done by "updating" those services.

```python
MAIN_SUBNETWORK = "main_subnetwork"
SECONDARY_SUBNETWORK = "secondary_subnetwork"

def run(args):
    # ...
    # all the services are added using the code snippet above
    # ...

    update_service(
        name="cassandra_node_2",
        config=UpdateServiceConfig(
            subnetwork=SECONDARY_SUBNETWORK,
        ),
    )
    update_service(
        name="app_service_2",
        config=UpdateServiceConfig(
            subnetwork=SECONDARY_SUBNETWORK,
        ),
    )
```

Kurtosis has the concept of a *default connection* between subnetworks. The default connection is the connection configuration that all pairs of subnetworks will inherit. At enclave creation, the default connection is set to allow communication between subnetworks. In this case, when the `cassandra_node_2` and `app_service_2` are assigned `secondary_subnetwork`, nothing changes. `main_subnetwork` and `secondary_subnetwork` can continue to communicate because the default connection is unblocked.

```
                     Kurtosis enclave                  
  -----------------------------------------------------
 |     main_subnetwork     |   secondary_subnetwork    |
 |-----------------------------------------------------|
 |                         .                           |
 |   cassandra_node_1      .                           |
 |                         .    cassandra_node_2       |
 |   cassandra_node_3      .                           |
 |    app_service_1        .                           |
 |                         .      app_service_2        |
 |   single_port_proxy     .                           |
 |                         .                           |
  -----------------------------------------------------
```

Setp 3: Configure the connection between the two subnetworks
------------------------------------------------------------

The creation of `main_subnetwork` and `secondary_subnetwork` now makes it possible to reconfigure the connection between them. Communications between the two subnetworks can now be blocked partially or completely by overriding the default connection for those two specific subnetworks.

```python
MAIN_SUBNETWORK = "main_subnetwork"
SECONDARY_SUBNETWORK = "secondary_subnetwork"

def run(args):
    # ...
    # all the services are added and `cassandra_node_2` and `app_service_2` are updated using the code snippet above
    # ...

    set_connection(
        (MAIN_SUBNETWORK, SECONDARY_SUBNETWORK), 
        config=ConnectionConfig(
            packet_loss_percentage=90.0, # 90% of the packet are dropped
        ),
    )
```

The result is the following: `cassandra_node_2` and `app_service_2` are partially unreachable and the resilience of the system can be tested.

```
                     Kurtosis enclave                  
  -----------------------------------------------------
 |     main_subnetwork     |   secondary_subnetwork    |
 |-------------------------|---------------------------|
 |                         |                           |
 |   cassandra_node_1      |                           |
 |                         |     cassandra_node_2      |
 |   cassandra_node_3      |                           |
 |    app_service_1        |                           |
 |                         |       app_service_2       |
 |   single_port_proxy     |                           |
 |                         |                           |
  -----------------------------------------------------
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->

[subnetworks-reference]: ../concepts-reference/subnetworks.md

[cap-theorem]: https://en.wikipedia.org/wiki/CAP_theorem

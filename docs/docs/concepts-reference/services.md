---
title: Services
sidebar_label: Services
---

A service is a user-created container running inside an enclave, which may or may not have ports exposed. 

They are added to an enclave with [Starlark](./starlark.md) using [`Plan.add_service`](../starlark-reference/plan.md#add_service), configured via [`ServiceConfig`](../starlark-reference/service-config.md), and removed from an enclave using [`Plan.remove_service`](../starlark-reference/plan.md#remove_service).

The services inside an enclave can be viewed with the [`kurtosis enclave inspect`](../cli-reference/enclave-inspect.md) CLI command.

For example:

```console
Name:            bold-rainforest
UUID:            04baa2402dd7
Status:          RUNNING
Creation Time:   Sat, 01 Apr 2023 17:13:32 -03

========================================= Files Artifacts =========================================
UUID           Name
6aaf173ee20c   indexer-config

========================================== User Services ==========================================
UUID           Name                        Ports                                      Status
7ebd3d73f03d   contract-helper-db          postgres: 5432/tcp -> 127.0.0.1:50223      RUNNING
349e86dc7f4a   contract-helper-dynamo-db   default: 8000/tcp -> 127.0.0.1:50228       RUNNING
efc406f91e29   contract-helper-service     rest: 3000/tcp -> http://127.0.0.1:8330    RUNNING
14508fc35534   explorer-backend            http: 8080/tcp -> http://127.0.0.1:18080   RUNNING
7c34bd0aba01   explorer-frontend           http: 3000/tcp -> http://127.0.0.1:8331    RUNNING
bb85411905d7   indexer-node                gossip: 24567/tcp -> 127.0.0.1:8333        RUNNING
                                           rpc: 3030/tcp -> http://127.0.0.1:8332
3a7db4c1e5b2   wallet                      http: 3004/tcp -> http://127.0.0.1:8334    RUNNING
```

Services, like other Kurtosis resources, are identified by [resource identifiers][resource-identifiers].

<!------------------ ONLY LINKS BELOW HERE -------------------->
[resource-identifiers]: ./resource-identifier.md

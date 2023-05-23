---
title: How Private & Public IPs & Ports Work
sidebar_label: How Private & Public IPs & Ports Work
---

The IP addresses and ports that come out of Kurtosis can be confusing at times. This document will explain how Kurtosis handles public and private IPs and ports.

### Private IPs
Each Docker or Kubernetes cluster has a private IP address range, from `0.0.0.0` to `255.255.255.255`. Each container gets IP addresses from this range, and containers talk to other containers using these IP addresses. These IP addresses have nothing to do with the outside world's IP addresses, and are purely internal to the cluster.

These IP addresses are private to Docker/Kubernetes, so you will not be able to reach them from your host machine (i.e. outside the Docker/Kubernetes cluster). For example, a container with IP address `172.0.1.3` will _not_ be reachable by `curl` from your regular command line.

### Private Ports
Containers can listen on ports. These ports do not conflict with other containers, nor do they conflict with ports on your host machine, outside the Docker/Kubernetes cluster. For example:

- You might be running Container A running a server listening on port `3000`, and Container B running a server listening on port `3000`. Even though the containers are using the same port, they are treated separately inside of Docker/Kubernetes and do not conflict.
- You might be running a server on your host machine on port `3000`, and a container in Docker/Kubernetes listening on port `3000`. This is also fine, because the Docker/Kubernetes ports are private to the cluster.

These container ports are private: you will not be able to access them from your host machine.

### Public IPs & Ports
To simplify your work, Kurtosis allows you to connect to every private port of every container. This is accomplished by binding every private port to an [ephemeral port](https://unix.stackexchange.com/questions/65475/ephemeral-port-what-is-it-and-what-does-it-do) on your host machine. These ephemeral ports are called the "public ports" of the container because they allow the container to be accessed outside the Docker/Kubernetes cluster. To view the private & public port bindings of each container in Kurtosis, run `kurtosis enclave inspect` and look for the bindings in the "Ports" column:

```
========================================== User Services ==========================================
GUID                     ID            Ports                                         Status
cl-beacon-1670597432     cl-beacon     http: 4000/tcp -> 127.0.0.1:55947             RUNNING
                                       metrics: 5054/tcp -> 127.0.0.1:55948
                                       tcp-discovery: 9000/tcp -> 127.0.0.1:55949
                                       udp-discovery: 9000/udp -> 127.0.0.1:52875
cl-validator-1670597459  cl-validator  http: 5042/tcp -> 127.0.0.1:55955             RUNNING
                                       metrics: 5064/tcp -> 127.0.0.1:55954
el-1670597405            el            engine-rpc: 8551/tcp -> 127.0.0.1:55930       RUNNING
                                       rpc: 8545/tcp -> 127.0.0.1:55928
                                       tcp-discovery: 30303/tcp -> 127.0.0.1:55927
                                       udp-discovery: 30303/udp -> 127.0.0.1:57433
                                       ws: 8546/tcp -> 127.0.0.1:55929
forkmon-1670597469       forkmon       http: 80/tcp -> 127.0.0.1:55962               RUNNING
grafana-1670597488       grafana       http: 3000/tcp -> 127.0.0.1:55998             RUNNING
```

The IP address used to reach these containers is your localhost address, `127.0.0.1`. This is the "public IP address" of each container in the cluster.

The combination of public IP + port _will_ allow you to connect to a container from your command line. For example, from the output above, `curl 127.0.0.1:55947` on your command line would make a request to private port `4000` on the `cl-client-0-beacon` service.

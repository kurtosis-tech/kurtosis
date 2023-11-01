---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---

[Kurtosis](https://github.com/kurtosis-tech/kurtosis) is a tool for packaging and launching environments of containerized services where you want them, the way you want them, with one liners.

- Get started with a quickstart to launch an environment.
- Learn about basic Kurtosis concepts.
- See our most popular use cases and explore real-world examples.

### Simple 1-line local deploy

```
kurtosis run github.com/kurtosis-tech/basic-service-package
```

<details><summary>**Result**</summary>

*CLI Output*

![basic-service-default-output.png](/img/home/basic-service-default-output.png)

*Service C UI, mapped locally*

![service-c-default.png](/img/home/service-c-default.png)
 
</details>

### 1-line deploy with feature flag and different numbers of each service

```
kurtosis run github.com/galenmarchetti/kurtosis-tech \
  '{"service_a_count": 2, 
    "service_b_count": 2, 
    "service_c_count": 1,
    "party_mode": true}'
```

<details><summary>**Result**</summary>

*CLI Output*

![basic-service-modified-cli-output.png](/img/home/basic-service-modified-cli-output.png)

*Service C UI, mapped locally*

![service-c-partying.png](/img/home/service-c-partying.png)
 
</details>
---
title: Introduction
sidebar_label: Introduction
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---

[Kurtosis](https://www.kurtosis.com) is a platform for orchestrating distributed system environments, allowing easy creation and manipulation of stage-appropriate deployments across the early stages of the development cycle (prototyping, testing).

Use cases for Kurtosis include:

* Enable individual developers to prototype on personal development environments without bothering with environment setup and configuration
* Enable development teams to run automated end-to-end tests for distributed systems, including fault tolerance tests in servers and networks, load testing, performance testing, etc.
* Enable developers to easily debug failing distributed systems during development

Why Kurtosis matters?
---------------------

Container management and container orchestration systems like Docker and Kubernetes are each great at serving developers in different parts of the development cycle (development for Docker, production for Kubernetes). These, and other distributed system deployment tools, are low-level, stage-specific tools that require teams of DevOps engineers to manage.

Kurtosis is designed to optimize **environment** management and control across the development cycle - operating at one level of abstraction higher than existing tools, giving developers the environments and the ability to manipulate them as needed at each stage.

![Why Kurtosis](../static/img/home/kurtosis-utility.png)

With Kurtosis, developers can build with local sandbox environments that demonstrate how their code will work when integrated with the rest of the system. In addition, advanced end-to-end testing workflows are available to teams using the manipulation tooling in the Kurtosis engine runtime which allow them to do end-to-end testing like fault-tolerance, regression, and performance tests.

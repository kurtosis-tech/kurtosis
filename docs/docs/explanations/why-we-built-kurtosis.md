---
title: Why We Built Kurtosis
sidebar_label: Why We Built Kurtosis
sidebar_position: 1
---

Kurtosis is a distributed application development platform. Its goal is to make building a distributed application as easy as developing a single-server app.

Our philosophy is that the distributed nature of modern software means that modern software development now happens at the environment level. Spinning up a single service container in isolation is difficult because it has implicit dependencies on other resources: services, volume data, secrets, certificates, and network rules. Therefore, the environment - not the container - is the fundamental unit of modern software.

This fact becomes apparent when we look at the software development lifecycle. Developers used to write code on their machine and ship a large binary to a few long-lived, difficult-to-maintain environments like Prod or Staging. Now, the decline of on-prem hardware, rise of containerization, and availability of flexible cloud compute enable the many environments of today: Prod, pre-Prod, Staging, Dev, and even ephemeral preview, CI test, and local dev.

The problem is that our tools are woefully outdated. The term "DevOps" was coined during the Agile revolution in the early 2000's. It signified making Dev responsible for end-to-end software delivery, rather than building software and throwing it over the wall to Ops to run. The idea was to shorten feedback loops, and it worked. However, our systems have become so complex that companies are now hiring "DevOps engineers" to manage the Docker, AWS, Terraform, and Kubernetes underlying all modern software. Though we call them "DevOps engineers", we are recreating Ops and separating Dev and Ops once more. 

Docker and Kubernetes are each great at serving developers in different parts of the development cycle: Docker for development/testing, Kubernetes for production. However, the separation between the two entails different distributed app definitions, and different tooling. In dev/test, this means Docker Compose and Docker observability tooling. In production, this means Helm definitions and manually-configured observability tools like Istio, Datadog, or Honeycomb.

Kurtosis aims to be at one level of abstraction higher. 

![Why Kurtosis](/img/home/kurtosis-utility.png)

In our vision, a developer should have a single platform for prototyping, testing, debugging, deploying to Prod, and observing the live system. Our goal with Kurtosis is to bring DevOps back.

To read more about our beliefs on reusable environments, [go here][reusable-environment-definitions]. To get started using Kurtosis, see [the quickstart][quickstart].

[reusable-environment-definitions]: ./reusable-environment-definitions.md
[quickstart]: ../quickstart.md

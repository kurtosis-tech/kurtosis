---
title: What Is Kurtosis?
sidebar_label: What Is Kurtosis?
sidebar_position: 1
---

Kurtosis is a new type of tool - what we're calling an "environment engine" - designed to make building a distributed system as easy as developing a single-server app.

Our philosophy is that the distributed nature of modern software means that modern software development now happens at the environment level. Spinning up a single service container in isolation is difficult because it has implicit dependencies on other resources like services, volume data, secrets, certificates, and network rules. Therefore, the environment - not the container - is the fundamental unit of modern software.

This fact becomes apparent when we look at the software development lifecycle. Developers used to write code on their machine and ship a large binary to a few long-lived, difficult-to-maintain environment like Prod or Staging. Now, the decline of on-prem hardware, rise of containerization, and availability of flexible cloud compute enable the many environments of today: Prod, pre-Prod, Staging, Dev, and even ephemeral preview, CI test, and local dev.

The problem is that our tools are woefully outdated. The term "DevOps" was coined during the Agile revolution in the early 2000's. It signified making Dev responsible for end-to-end software delivery, rather than building software and throwing it over the wall to Ops to run. The idea was to shorten feedback loops, and it worked. However, our systems have become so complex that companies are now hiring "DevOps engineers" to manage the Docker, AWS, Terraform, and Kubernetes underlying all modern software. Though we call them "DevOps engineers", we are recreating Ops and separating Dev and Ops once more. 

In our vision, a developer should have a single tool for their service that can prototype, test, debug, deploy to Prod, and observe while running live. Our goal with Kurtosis is to bring DevOps back. 

To read more about our beliefs on reusable environments, [go here][reusable-environment-definitions]. To get started using Kurtosis, see [the installation instructions][install].

[reusable-environment-definitions]: ./reusable-environment-definitions.md
[install]: ../guides/installing-the-cli.md

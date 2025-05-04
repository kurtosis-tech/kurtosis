---
title: Roadmap
sidebar_label: Roadmap
slug: '/roadmap'
---

::: note:
Last updated: May 1, 2025
:::

:::note
Kurtosis Technologies open sourced Kurtosis in June '24. Since then, Kurtosis has grown via open source contributions and active maintenance. If you have ideas to improve Kurtosis, please make a PR to suggest them under Kurtosis Improvement Proposals and get in touch with one of the [maintainers].
:::

Over the next 3-6 months, Kurtosis maintainers are looking to improve the following product areas:

**Persisting enclave data** 

Enclaves hold valuable state that can be used for debugging, reproducing environments, and saving time getting the environments you need spun up. Currently, Kurtosis stores a lot of information about the state of the enclave (e.g. service information, files artifacts, persistent directories, running conatiners), but there's easy way to extract this information for reproducing the environment.

We'll be exploring ways to allow users to restart enclaves, snapshot enclave states, plug data into enclaves, etc.

Please reach out to [Tedi Mitiku](https://tedi.dev) if any of this is relevant to you/your teams workflows.

**Faster local development loop**

With the ecosystem of Kurtosis packages growing, packages are now being composed together to form larger packages - luckily Kurtosis makes composing packages very easy. The downside is that packages are getting bigger and bigger, and taking longer and longer to spin up on users machines for testing. Once a package is spun up and the enclave is running, developers want to iterate by make changes to code in their services and then see those changes reflected in, make changes to the enclave quickly.

 
**Support for longer term K8s environments** that enable a developer to make changes to the Starlark package & Kurtosis will apply those changes to a long-lived enclave deterministically.
**

If any of the investments we are making interest you or if you have feedback for us, please let us know in our [Github Discussions](https://github.com/kurtosis-tech/kurtosis/discussions/categories/q-a), [Discord](https://discord.com/invite/TMhR2uX5WMZ), or reach out to maintainers and/or [Tedi Mitiku](https://tedi.dev) to chat more.

### Kurtosis Improvement Proposals

Kurtosis encourages users to fork Kurtosis and make improvements to the engine as they see fit. Kurtosis also encourages contributions and proposals for improvements for larger features requiring consideration and coordination. If you are interested in proposing features, please create a doc and make a PR to add it to this list!
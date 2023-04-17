---
title: feedback
sidebar_label: feedback
slug: /feedback
---

To quickly get in touch with us, simply run:
```bash
kurtosis feedback
```
which will open the [Kurtosis GitHub create new issue page](https://github.com/kurtosis-tech/kurtosis/issues/new/choose), where you can file a bug report, submit a feature request, or let us know about any docs issues you may have come across. A link to our [Discord server](https://discord.gg/yUGgeE8s) will be made available there too.

### Using `kurtosis feedback`

The `kurtosis feedback` command can accept both an argument and various flags. To send our team feedback over GitHub straight from the CLI, simply pass in your feedback as an argument to the command like so:
```bash
kurtosis feedback [flags] "$YOUR_FEEDBACK"
```
where `YOUR_FEEDBACK` is the feedback you would like to send to us. 

Running just `kurtosis feedback "my feedback"` (with no flags) will open the new issue creation page on [our GitHub](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) where you can can select the issue type and have the description field pre-filled with `my feedback`. 

Below are a collection of valid flags you may use: 
- The `--email` flag opens a draft email to our team, via your default mail client, that has the body of the email pre-filled with whatever you entered in the `"$YOUR_FEEDBACK"` arg. This is the default behavior when no flag is set.
- The `--calendly` flag can be used to open our [Calendly link](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding) to schedule time with our team to help you get started, address feedback, and answer any questions you may have.
Additionally, the flags below can be used alone (will set the feedback type for GitHub which is the default feedback destination) and with the `--email` flag to specify the *type* of feedback you wish to provide:
- The `--bug` flag can be used when you wish to submit a bug report to us. When this `--bug` flag is set without any destination flag, the CLI will take you directly to the [bug report issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=bug&template=bug-report.yml) in our GitHub. When this `--bug` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[BUG]`, which will help our team triage and prioritize your report.
- The `--feature` flag can be used when you wish to submit a feature request to us. When this `--feature` flag without any destination flag, the CLI will take you directly to the [feature request issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=feature+request&template=feature-request.yml) in our GitHub. When this `--feature` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[FEATURE_REQUEST]`, which will help our team triage and prioritize your request.
- The `--docs` flag can be used when you wish to flag an issue with our documentation. When this `--docs` flag without any destination flag, the CLI will take you directly to the [docs issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=leeederek&labels=docs&template=docs-issue.yml) in our GitHub. When this `--docs` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[DOCS]`, which will help our team triage and prioritize the issue.

:::tip
To join our Discord community, use the [`kurtosis discord`](./discord.md) CLI command.
:::

---
title: feedback
sidebar_label: feedback
slug: /feedback
---

Your feedback is valuable and helps us improve Kurtosis. We thank you in advance for taking the time to share your suggestions and concerns with us.

To send our team feedback over email straight from the CLI, simply run:
```bash
kurtosis feedback "$YOUR_FEEDBACK"
```
where `$YOUR_FEEDBACK` is the feedback you would like to send to us. For example, `$YOUR_FEEDBACK` could be: "_I enjoy the enclave names_".

Running the above command will open a draft email using your default email client with the `TO:` line pre-filled with our feedback email list (feedback@kurtosistech.com) and the body of the message pre-filled with `$YOUR_FEEDBACK` (that you entered in the command line).

To quickly see all the ways you can get in touch with us, simply run:
```bash
kurtosis feedback
```

which will return:
- The link to our [Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) for filing bug reports and feature requests as Github Issues. 
- Our [feedback email address](mailto:feedback@kurtosistech.com) for sending us suggestions, comments, or questions about Kurtosis via email.
- A [Calendly link](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding) to schedule a dedicated 1:1 onboarding session with us to help you get started.

##### Using `kurtosis feedback`

The `kurtosis feedback` command can accept both an argument and various flags. To send our team feedback over Github straight from the CLI, simply pass in your feedback as an argument to the command like so:
```bash
kurtosis feedback [flags] ["YOUR_FEEDBACK"]
```
where `YOUR_FEEDBACK` is the feedback you would like to send to us. 

Running just `kurtosis feedback "my feedback"` (with no flags) will open the new issue creation page on [our Github](https://github.com/kurtosis-tech/kurtosis/issues/new/choose) where you can can select the issue type and have the description field pre-filled with `my feedback`. 

Below are a collection of valid flags you may use:
- The `--github` flag can be used to open the Issue creation page in our Github where you can select the Issue template you wish to use for your feedback. The `"$YOUR_FEEDBACK"` arg will be pre-populated in the description of whichever Issue template you select.
- The `--email` flag opens a draft email to feedback@kurtosistech.com, via your default mail client, that has the body of the email pre-filled with whatever you entered in the `"$YOUR_FEEDBACK"` arg. This is the default behavior when no flag is set.
- The `--calendly` flag can be used to open our [Calendly link](https://calendly.com/d/zgt-f2c-66p/kurtosis-onboarding) to schedule time with our team to help you get started, address feedback, and answer any questions you may have.
Additionally, the flags below can be used with `--github` and `--email` to specify the *type* of feedback you wish to provide:
- The `--bug` flag can be used when you wish to submit a bug report to us. When this `--bug` flag is set alongside the `--github` flag, the CLI will take you directly to the [bug report issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=bug&template=bug-report.yml&title=%5Bbug%5D%3A+) in our Github. When this `--bug` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[BUG]`, which will help our team triage and prioritize your report.
- The `--fr` flag can be used when you wish to submit a feature request to us. When this `--fr` flag is set alongside the `--github` flag, the CLI will take you directly to the [feature request issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=feature+request&template=feature-request.yml&title=%5BFR%5D%3A+) in our Github. When this `--fr` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[FEATURE_REQUEST]`, which will help our team triage and prioritize your request.
- The `--docs` flag can be used when you wish to flag an issue with our documentation. When this `--docs` flag is set alongside the `--github` flag, the CLI will take you directly to the [docs issue creation page](https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=leeederek&labels=docs&template=docs-issue.yml&title=%5BDocs%5D%3A+) in our Github. When this `--docs` flag is set alongside the `--email` flag, the CLI will open an email draft with the subject pre-filled with: `[DOCS]`, which will help our team triage and prioritize the issue.

:::tip
To join our Discord community, use the [`kurtosis discord`](./discord.md) CLI command.
:::

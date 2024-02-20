---
title: github login
sidebar_label: github login
slug: /github-login
---

To authorize Kurtosis CLI to GitHub, use: 

```console
kurtosis github login
```

This enables use of any private GitHub assets that a user has access via a [locator](../advanced-concepts/locators.md) in operations like `kurtosis run`, `import_module`, or `upload_files`. To see an application of this, follow this [guide](../guides/running-private-packages.mdnn) to learn how to run a private package.

The command will output a one time code. Copy the code and press enter to open a GitHub window that will instruct you to enter the code. After entering the code, authorize Kurtosis CLI and navigate back to the terminal. Your Kurtosis engine will restart automatically so GitHub auth takes effect.

Under the hood, GitHub will provide Kurtosis CLI with a restricted OAuth token that authorizes Kurtosis CLI to perform GitHub operations on your behalf, such as reading a private repository. Kurtosis follows the same pattern as [GitHub CLI](https://cli.github.com/manual/gh_auth_login) and stores the token in secure system credential storage. If a system credential store is not detected, the token is stored in a plain text file in the Kurtosis config directory at `kurtosis config path`.
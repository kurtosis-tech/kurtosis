---
title: Best Practices
sidebar_label: Best Practices
slug: /best-practices
---

Passing package arguments to the CLI
-------------------------------
Passing [package parameters][package-parameterization] via the CLI can get hairy due to the interaction between Bash and JSON quotes. The following are tips to make your life easier:

1. **When you have a small number of arguments:** surround the arguments with single quotes so you don't have to escape double quotes in your JSON. E.g.:
   ```bash
   kurtosis run github.com/user/repo '{"some_param":5,"some_other_param":"My value"}'
   ```
1. **When you have a large number of arguments:** put them in a `.yaml` or `.json` file and use `--args-file` to slot them into the `kurtosis run` command. E.g.:
   ```bash
   kurtosis run github.com/user/repo --args-file my-params.yaml
   ```
   or if you prefer JSON:
   ```bash
   kurtosis run github.com/user/repo --args-file my-params.json
   ```

Choosing the right wait
-----------------------
Kurtosis has three different types of waits. Described here are the three, with tips on when to use each:

1. Automatic waiting on port availability when a service starts (enabled by default; can be configured with [`PortSpec.wait`][port-spec-starlark-reference])
    - Should be sufficient for most usecases
    - Requires little-to-no extra configuration
    - Will cause parallel `Plan.add_services` to fail, allowing for quick aborting
1. Waiting on [`ReadyCondition`][ready-condition-starlark-reference]s (configured in [`ServiceConfig`][service-config-starlark-reference])
    - Allows for more advanced checking (e.g. require a certain HTTP response body, ensure a CLI call returns success, etc.)
    - More complex to configure
    - Will cause parallel `Plan.add_services` to fail, allowing for quick aborting
1. The [`Plan.wait`][plan-wait-starlark-reference]
    - Most useful for asserting the system has reached a desired state in tests (e.g. wait until data shows up after loading)
    - More complex to configure
    - Cannot be used to short-circuit `Plan.add_services`

Choosing service names in Kurtosis
----------------------------------
Kurtosis service names implements [RFC-1035](https://datatracker.ietf.org/doc/html/rfc1035), meaning the names of all services must be a valid [RFC-1035 Label Name](https://kubernetes.io/docs/advanced-concepts/overview/working-with-objects/names/#rfc-1035-label-names). Tactically this means a service name must:

- contain at most 63 characters
- contain only lowercase alphanumeric characters or '-'
- start with an alphabetic character
- end with an alphanumeric character

Failure to adhere to the above standards will result in errors when running Kurtosis.

Writing and reading Starlark
----------------------------

If you're using Visual Studio Code, you may find our [Kurtosis VS Code Extension][vscode-plugin] helpful when writing Starlark.

If you're using Vim, you can add the following to your `.vimrc` to get Starlark syntax highlighting:

```
" Add syntax highlighting for Starlark files
autocmd FileType *.star setlocal filetype=python
```

or if you use Neovim:
```
autocmd BufNewFile,BufRead *.star set filetype=python
```

<!---------------------------------------- ONLY LINKS BELOW HERE!!! ----------------------------------->
[package-parameterization]: ./advanced-concepts/packages.md#parameterization
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
[service-config-starlark-reference]: ./api-reference/starlark-reference/service-config.md
[port-spec-starlark-reference]: ./api-reference/starlark-reference/port-spec.md
[ready-condition-starlark-reference]: ./api-reference/starlark-reference/ready-condition.md
[plan-wait-starlark-reference]: ./api-reference/starlark-reference/plan.md#wait

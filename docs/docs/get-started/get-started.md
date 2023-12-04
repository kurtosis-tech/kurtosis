---
id: get-started
title: Kurtosis Documentation
slug: '/'
sidebar_position: 1
hide_table_of_contents: true
---

[Kurtosis](https://github.com/kurtosis-tech/kurtosis) is a tool for packaging and launching environments of containerized services where you want them, the way you want them, with one liners.

- Get started with a [quickstart](quickstart.md) to launch an environment.
- Dive deeper with [basic Kurtosis concepts](basic-concepts.md).
- Write your own environment definition with a guide on [writing a package](write-your-first-package.md).

To quickly see what Kurtosis feels like, check out the example snippets below:

### Local Deploy on Docker

```console
kurtosis run github.com/kurtosis-tech/basic-service-package
```

<details><summary><b>Output</b></summary>

*CLI Output*

![basic-service-default-output.png](/img/home/basic-service-default-output.png)

*Example Service C UI, mapped locally*

![service-c-default.png](/img/home/service-c-default.png)
 
</details>

### Local deploy with feature flag and different numbers of each service

```console
kurtosis run github.com/kurtosis-tech/basic-service-package \
  '{"service_a_count": 2, 
    "service_b_count": 2, 
    "service_c_count": 1,
    "party_mode": true}'
```

<details><summary><b>Output</b></summary>

*CLI Output*

![basic-service-modified-cli-output.png](/img/home/basic-service-modified-cli-output.png)

*Example Service C UI, mapped locally*

![service-c-partying.png](/img/home/service-c-partying.png)
 
</details>

### Remote deploy on Kubernetes

```console
kurtosis cluster set remote-kubernetes; kurtosis gateway > /dev/null 2>&1 &
```
```console
kurtosis run github.com/kurtosis-tech/basic-service-package \
  '{"service_a_count": 2, 
    "service_b_count": 2, 
    "service_c_count": 1,
    "party_mode": true}'
```

<details><summary><b>Output</b></summary>

**Note:** The experience on remote k8s is the same as local Docker.

*CLI Output*

![basic-service-modified-cli-output.png](/img/home/basic-service-output-k8s.png)

*Example Service C UI, mapped locally*

![service-c-partying.png](/img/home/service-c-k8s.png)
 
</details>

Writing and reading Starlark
----------------------------
The Starlark programming language is used to manipulate the contents of environments spun up with Kurtosis. Read more about Starlark and why we chose to use it [here](../advanced-concepts/why-kurtosis-starlark.md).

If you're using Visual Studio Code to write and edit Starlark in Kurtosis packages, you may find our [Kurtosis VS Code Extension][vscode-plugin] helpful when writing Starlark.

If you're using Vim, you can add the following to your `.vimrc` to get Starlark syntax highlighting:

```
" Add syntax highlighting for Starlark files
autocmd FileType *.star setlocal filetype=python
```

or if you use Neovim:
```
autocmd BufNewFile,BufRead *.star set filetype=python
```

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[vscode-plugin]: https://marketplace.visualstudio.com/items?itemName=Kurtosis.kurtosis-extension
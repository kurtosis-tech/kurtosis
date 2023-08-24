---
title: Adding Command-Line Completion
sidebar_label: Adding Command-Line Completion
sidebar_position: 4
---

<!-- NOTE TO KURTOSIS DEVS: 

This page was created by referencing the kubectl docs:
* https://kubernetes.io/docs/tasks/tools/included/optional-kubectl-configs-bash-linux/
* https://kubernetes.io/docs/tasks/tools/included/optional-kubectl-configs-bash-mac/
* https://kubernetes.io/docs/tasks/tools/included/optional-kubectl-configs-zsh/
* https://kubernetes.io/docs/tasks/tools/included/optional-kubectl-configs-fish/

-->

<!---------- START IMPORTS ------------>

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<!---------- END IMPORTS ------------>

[The Kurtosis CLI](../cli-reference/index.md) supports command-line completion for `bash`, `zsh`, and `fish`. With completion installed, you will be able to:

- Complete subcommands (e.g. typing `kurtosis` and pressing TAB will suggest subcommands)
- Complete dynamic arguments (e.g. typing `kurtosis enclave inspect` and pressing TAB will list the names of existing enclaves if any)

The process for installing completion is specific to each shell:

<Tabs groupId="install-methods">
<TabItem value="bash" label="bash">

1. Print your Bash version:
    ```bash
    bash --version
    ```
1. If your Bash version is less than 4.1, upgrade it:
    * On Mac, upgrade Mac via Homebrew:
        ```bash
        brew install bash
        ```
    * On Linux, [upgrade it via the package manager for your distro](https://www.configserverfirewall.com/linux-tutorials/update-bash-linux/)
1. Check if you have [bash-completion](https://github.com/scop/bash-completion) installed:
    ```bash
    type _init_completion
    ```
1. If you get an error like `-bash: type: _init_completion: not found`, install Bash completion:
    * On Mac:
        1. Install the completion library:
            ```bash
            brew install bash-completion@2
            ```
        1. Add the following to your `~/.bash_profile`:
            ```bash
            export BREW_PREFIX="$(brew --prefix)"
            [[ -r "${BREW_PREFIX}/etc/profile.d/bash_completion.sh" ]] && source "${BREW_PREFIX}/etc/profile.d/bash_completion.sh"
            ```
        1. Close and re-open your terminal window to reload your shell.
        1. Verify that you now have the completion installed:
            ```bash
            type _init_completion
            ```
    * On Linux, install it using the package manager for your distro using [these installation instructions](https://github.com/scop/bash-completion#installation)
1. Skip this step if you are installing using Homebrew and have `bash-completion@2` installed. Otherwise, proceed to source the output of `kurtosis completion bash` in your Bash config file:
    * On Mac, add the following to your `~/.bash_profile` file:
        ```bash
        # Add Kurtosis command-line completion
        source <(kurtosis completion bash)
        ```
    * On Linux, add the following to your `~/.bashrc` file:
        ```bash
        # Add Kurtosis command-line completion
        source <(kurtosis completion bash)
        ```
1. If you have an alias set up for Kurtosis, add completion for that as well (we'll assume the alias `kt` in the examples below):
    * On Mac, add the following to your `~/.bash_profile` file:
        ```bash
        # Add command-line completion to Kurtosis alias
        complete -F __start_kurtosis kt
        ```
    * On Linux, add the following to your `~/.bashrc` file:
        ```bash
        # Add command-line completion to Kurtosis alias
        complete -F __start_kurtosis kt
        ```
1. Close and re-open your terminal window to reload your shell and apply the changes.

</TabItem>

<TabItem value="zsh" label="zsh">

1. Add the following to your `~/.zshrc` file:
    ```zsh
    # Add Kurtosis command-line completion
    source <(kurtosis completion zsh)
    compdef _kurtosis kurtosis
    ```
1. If you have an alias set up for Kurtosis, add the following to your `~/.zshrc` file (we'll assume the alias `kt` in this example):
    ```zsh
    # Add command-line completion to Kurtosis alias
    compdef __start_kurtosis kt
    ```
1. Close and re-open your terminal window to reload your shell and apply the changes.
1. If you get an error like `complete:13: command not found: compdef`, add the following to the top of your `~/.zshrc` and close and re-open your terminal window to reload your shell:
    ```zsh
    autoload -Uz compinit
    compinit
    ```

</TabItem>
<TabItem value="fish" label="fish">

1. Add the following to your `~/.config/fish/config.fish` file:
    ```fish
    # Add Kurtosis command-line completion
    kurtosis completion fish | source
    ```
1. Close and re-open your terminal window to reload your shell and apply the changes.

</TabItem>
<TabItem value="manual" label="Manual Installation">

If necessary, tab completion can be installed manually in two steps as follows, by first generating the 
tab completion code (specific to the shell) and then sourcing that code into the shell. 

1. The code needed to enable tab completion can be generated by the `kurtosis` cli by
   running `kurtosis completion <SHELL>` command, e.g. for `bash`:
   ```
   kurtosis completion bash 
   ```

1. `source`ing the output of the command will enable command-line completion, and adding the `source`
   command to your shell config file will enable it across shell instances.
   ```
   # Add Kurtosis command-line completion to your shell config file
   source <(kurtosis completion bash)
   ```

</TabItem>
</Tabs>

---
title: Best Practices
sidebar_label: Best Practices
slug: /best-practices
---

Passing package arguments to the CLI
-------------------------------
Passing [package arguments][args-concepts-reference] to the CLI can get hairy due to the interaction between Bash and JSON quotes. The following are tips to make your life easier:

1. **When you have a small number of arguments:** surround the arguments with single quotes so you don't have to escape double quotes in your JSON. E.g.:
   ```bash
   kurtosis run github.com/user/repo '{"some_param":5,"some_other_param":"My value"}'
   ```
1. **When you have a large number of arguments:** put them in a `.json` file and use [Bash command substitution](https://www.gnu.org/software/bash/manual/html_node/Command-Substitution.html) _inside double quotes_ to slot them into the `kurtosis run` command. E.g.:
   ```bash
   kurtosis run github.com/user/repo "$(cat my-args.json)"
   ```
   The double quotes around the `$(cat my-args.json)` are important so any spaces inside `my-args.json` don't fool Bash into thinking you're passing in two separate arguments.


<!---------------------------------------- ONLY LINKS BELOW HERE!!! ----------------------------------->
[args-concepts-reference]: ./concepts-reference/args.md

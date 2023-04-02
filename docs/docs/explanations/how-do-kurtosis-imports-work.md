---
title: How Do Kurtosis Imports Work?
sidebar_label: How Do Kurtosis Imports Work?
---

### Background
Kurtosis allows a [Starlark script][starlark-reference] to use content from other files. This might be importing code from another Starlark file (via [the `import_module` instruction][import-module-starlark-reference]), or using contents of a static file (via [the `read_file` instruction][read-file-starlark-reference]).

In both cases, the Kurtosis engine needs to know where to find the external file. There are two cases where external files might live:

1. Locally: the external file lives on the same filesystem as the Starlark script that is trying to use it.
1. Remotely: the external file lives somewhere on the internet.

Therefore, Kurtosis needs to handle both of these.

:::info
This external files problem is not unique to Kurtosis. Every programming language faces the same challenge, and each programming language solves it differently. 

<details>
<summary>Click to see examples</summary>

- In Javascript, local files are referenced via relative imports:

  ```javascript
  import something from ../../someDirectory/someFile
  ```

  and remote files are downloaded as modules using `npm` or `yarn` and stored in the `node_modules` directory. The remote files will then be available via:

  ```javascript
  import some-package
  ```

- In Python, local files are handled via the [relative import syntax](https://docs.python.org/3/reference/import.html#package-relative-imports):

  ```python
  from .moduleY import spam
  from ..moduleA import foo
  ```

  and remote files are downloaded as packages using `pip`, stored somewhere on your machine, and made available via the `PYTHONPATH` variable. The package will then be available via regular import syntax:

  ```python
  import some_package
  ```

- In Java, the difference between local and remote files is less distinct because all files are packaged in JARs. Classes are imported using Java's import syntax:

  ```java
  import com.docker.clients.Client;
  ```

  and the Java classpath is searched for each import to see if any JAR contains a matching file. It is the responsibility of the user to build the correct classpath, and various tools and dependency managers help developers download JARs and construct the classpath correctly.

</details>
:::

### Kurtosis Packages
Remote file imports in any language are always handled through a packaging system. This is because any language that allows remote external files must have a solution for identifying the remote files, downloading them locally, and making them available on the import path (`PYTHONPATH`, `node_modules`, classpath, etc.). Furthermore, authors must be able to bundle files together into a package, publish them, and share them. Thus, For Kurtosis to allow Starlark scripts to depend on remote external files, we needed a packaging system of our own.

Of all the languages, we have been most impressed by [Go's packaging system (which Go calls "modules")](https://go.dev/blog/using-go-modules). In Go:

- Modules are easy to create by adding a `go.mod` manifest file to a directory ([example](https://github.com/kurtosis-tech/kurtosis/blob/main/cli/cli/go.mod))
- Dependencies are easy to declare in the `go.mod` file
- Modules are published to the world simply by pushing up to GitHub

Kurtosis code needs to be easy to share, so we modelled our packaging system off Go's.

In Kurtosis, a directory that has [a `kurtosis.yml` file][kurtosis-yml-reference] is the package root of a [Kurtosis package][packages-reference], and all the contents of that directory will be part of the package. Any Starlark script inside the package will have the ability to use external files (e.g. via `read_file` or `import_module`) by specifying [the locator][locators-reference] of the file.

Each package will be named with the `name` key inside the `kurtosis.yml` file. Package names follow the format `github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml` as specified in [the `kurtosis.yml` documentation][kurtosis-yml-reference]. This package name is used to determine whether a file being imported is local (meaning "found inside the package") or remote (meaning "found from the internet"). The logic for resolving a `read_file`/`import_module` is as follows:

- If the package name in the `kurtosis.yml` is a prefix of the [locator][locators-reference] used in `read_file`/`import_module`, then the file is assumed to be local inside the package. The package name in the locator (`github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml`) references the package root (which is the directory where the `kurtosis.yml` lives), and each subpath appended to the package name will traverse down in the repo.

- If the package name is not a prefix of the [locator][locators-reference] used in `read_file`/`import_module`, then the file is assumed to be remote. Kurtosis will look at the `github.com/package-author/package-repo` prefix of the locator, clone the repository from GitHub, and use the file inside the package i.e a directory that contains kurtosis.yml. 

:::info
Since `kurtosis.yml` can live in any directory, users have the ability to create multiple packages per repo (sibling packages). We do not currently support a package importing a sibling package (i.e. if `foo` and `bar` packages are subdirectories of `repo`, then `bar` cannot import files from `foo`). Please let us know if you need this functionality.
:::

Kurtosis does not allow referencing local files outside the package (i.e. in a directory above the package root with the `kurtosis.yml` file). This is to ensure that all files used in the package get pushed to GitHub when the package is published.

### Packages in Practice
There are three ways to run Kurtosis Starlark. The first is by running a script directly:

```
kurtosis run some-script.star
```

Because only a script was specified, Kurtosis does not have the `kurtosis.yml` or package name necessary to resolve file imports. Therefore, any imports used in the script will fail.

The second way is to run a runnable package by pointing to the package root:

```
# OPTION 1: Point to the directory containing the `kurtosis.yml` and `main.star`
kurtosis run /path/to/package/root   # Can also be "."

# OPTION 2: Point to a `kurtosis.yml` file directly, with a `main.star` next to it
kurtosis run /path/to/package/root/kurtosis.yml
```

In both cases, Kurtosis will run the `main.star` in the package root and resolve any file imports using the package name specified in the `kurtosis.yml`. All local imports (imports that have the package name as a prefix to the locator) will be resolved within the directory on your filesystem; this is very useful for local development.

:::info
Not all packages have a `main.star` file, meaning not all packages are runnable; some packages are simply libraries intended to be imported in other packages.
:::

The third way is to run a runnable package by its package name (can be found in the kurtosis.yml from the directory):

```
# if kurtosis.yml is in repository root
kurtosis run github.com/package-author/package-repo
```

```
# if kurtosis.yml is in any other directory
kurtosis run github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml
```

Kurtosis will clone the package from GitHub, run the `main.star`, and use the `kurtosis.yml` to resolve any imports. This method always uses the version on GitHub.

:::tip
If you want to run a non-main branch, tag or commit use the following syntax
`kurtosis run github.com/package-author/package-repo@tag-branch-commit`
:::

<!-- 
  It seems to me that we are suggesting users to use arbitrary name, only to change later; my worry is that it
  could lead to import errors! With introduction of sub-packages, this could lead to even more confusion. If the users'
  want to do this for quick testing, they can but we should not suggest it.
-->
:::tip
When you're developing locally, before your package has been pushed to GitHub, the package `name` can be anything you like - e.g. `github.com/test/test`. The only thing that is important for correctly resolving local file imports is that your `read_file`/`import_module` locators also are prefixed with `github.com/test/test`.

Once you push to GitHub, however, your package `name` will need to match the author and repo. If they don't, your package will be broken when another user depends on your package because Kurtosis will go looking for a `github.com/test/test` package that likely doesn't exist.
:::

<!---------------------- ONLY LINKS BELOW HERE ---------------------------->
[starlark-reference]: ../concepts-reference/starlark.md
[kurtosis-yml-reference]: ../concepts-reference/kurtosis-yml.md
[packages-reference]: ../concepts-reference/packages.md
[locators-reference]: ../concepts-reference/locators.md
[import-module-starlark-reference]: ../starlark-reference/import-module.md
[read-file-starlark-reference]: ../starlark-reference/read-file.md

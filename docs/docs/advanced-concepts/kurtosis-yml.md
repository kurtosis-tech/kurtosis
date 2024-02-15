---
title: kurtosis.yml
sidebar_label: kurtosis.yml
---

:::info
The `kurtosis.yml` is part of the Kurtosis package system. To read about the package system in detail, [see here][how-do-kurtosis-imports-work-explanation].
:::

The `kurtosis.yml` file is a manifest file necessary to turn a directory into [a Kurtosis package][package]. This is the spec for the `kurtosis.yml`:

<!-- TODO UPDATE THIS WHEN DEPENDENCIES GO HERE -->

```yaml
# The locator naming this package.
name: github.com/package-author/package-repo/path/to/directory-with-kurtosis.yml
# The package's description which will be shown in the Enclave Manager on the UI.
description: A sentence describing what the package does
# The package's dependencies replace options
replace:
  # Replacing the official Postgres package with my fork
  github.com/kurtosis-tech/postgres-package: github.com/my-github-user/postgres-package
```

Example usage:

if kurtosis.yml is in the repository root:
```yaml
name: github.com/author/package-repo
```

if kurtosis.yml is in a directory other than repository root:
```yaml
name: github.com/author/package-repo/path/to/directory-with-kurtosis.yml
```

:::info
The key take away is that `/path/to/directory-with-kurtosis.yml` only needs to be provided if `kurtosis.yml` is not present in the repository's root.
:::

Replace
-------
There are often times when you want to substitute one of your Kurtosis package’s dependencies with another dependency. For example, someone might have forked one of your package's dependencies, and you want to test your package against their fork rather than the normal version. Finding and updating all the dependency-referencing commands (`import_module` , `upload_file`, `read_file`, etc.) in your package is tedious and error-prone, so the `kurtosis.yml` supports a `replace` key to do it for you.

The `replace` key accepts a key-value map where each key is the [locator][locators] of a package to be replaced, and each value is the package locator to replace it with.

For example:

```yaml
name: github.com/my-github-user/my-package
replace:
  # Replacing the official Postgres package with my fork
  github.com/kurtosis-tech/postgres-package: github.com/my-github-user/postgres-package
```

This behaves just as if you’d manually updated each Starlark dependency-referencing command that consumes `github.com/kurtosis-tech/postgres-package` and replaced it with `github.com/my-github-user/postgres-package`. This replace includes transitive dependencies: a dependency package that itself uses `github.com/kurtosis-tech/postgres-package` will _also_ instead now use `github.com/my-github-user/postgres-package`!

:::info
A `replace` entry does nothing if replaced package isn't actually depended upon.
:::

You may optionally append an `@` and version after the replacement package locator to specify which version of the replacement package ought to be used.

For example:

```yaml
name: github.com/my-github-user/my-package
replace:
  # Replacing the official Postgres package with version 1.2.3
  github.com/kurtosis-tech/postgres-package: github.com/my-github-user/postgres-package@1.2.3
```

Like `import_module` and all other dependency-referencing commands, the version can be a tag, branch, or a full commit hash.

:::info
Go programmers will identify the similarities with the `replace` directive in the `go.mod` file. This is not accidental; the Kurtosis packaging system draws heavy inspiration from the Go module system.
:::

### Replace In kurtosis.yml Of Dependencies
`replace` instructions are only evaluated in the `kurtosis.yml` of the root package being called, and are ignored in the `kurtosis.yml`s of package dependencies.
For example, suppose we had two packages, Dependency and Consumer, that both use `github.com/kurtosis-tech/postgres-package` in their Starlark.
Additionally, Consumer depends on Dependency in its Starlark code.
Their `kurtosis.yml` files look like so:

Dependency:

```yaml
# NOTE: this package uses github.com/kurtosis-tech/postgres-package in its Starlark

name: github.com/somebody/dependency
replace:
  # Replace the official Postgres package with the fork from user 'someboday'
  github.com/kurtosis-tech/postgres-package: github.com/somebody/postgres-package
```

Consumer:

```yaml
# NOTE: this package uses both of these in its Starlark:
# - github.com/somebody/dependency
# - github.com/kurtosis-tech/postgres-package

name: github.com/somebody/consumer

# This package does NOT have a 'replace' directive
```

If the user runs Consumer (`kurtosis run github.com/somebody/consumer`), **Dependency's `replace` will not be evaluated**. This is because `replace` instructions are only executed in the `kurtosis.yml` of the root package.

### Local Paths
You might need to replace one of your package's GitHub dependencies with a local version on your filesystem. For example, suppose you're developing on one of your package's dependencies. When doing so, updating all the dependency-referencing commands like `import_module` would be painful. To support this use case, the value of a `replace` line can be an absolute or relative path on your local filesystem to a Kurtosis package (a directory containing a `kurtosis.yml` file).

For example:

```yaml
name: github.com/mieubrisse/my-package

replace:
    # Replace the official Postgres package with the version on my filesystem (relative import)
    github.com/kurtosis-tech/postgres-package: ../postgres-package

    # Replace the official MongoDB package with the version on my filesystem (absolute import)
    github.com/kurtosis-tech/mongodb-package: /home/code/mongodb-package
```

Kurtosis identifies a local package filepath whenever the replacement package locator begins with `/` or `.`. If the filepath does not point to a directory containing a `kurtosis.yml` file, an error will be thrown. When local filepath locators are used, any `@` in the value will be treated as part of the filepath (rather than indicating a version).

### Colliding Replace Values
It is possible to have two Kurtosis packages, one nested within the other.
For example, suppose we have `github.com/kurtosis-tech/parent` and `github.com/kurtosis-tech/parent/child`.
To the Kurtosis packaging system, these packages do not have any hierarchical relation and are simply treated as two entirely separate packages with different identifiers for dependency purposes.
However, this still presents a problem for `replace`: let's suppose you have your own package that depends on both `parent` and `child`, and you want to `replace` both.
How does this work?

When one package exists in a subpath of another, the more specific (child) package gets replaced first. For example, if your package's `kurtosis.yml` looks like so...

```yaml
name: github.com/mieubrisse/my-package
replace:
  # Replace parent with new-parent
  github.com/kurtosis-tech/parent: github.com/mieubrisse/new-parent

  # Replace child with new-child
  github.com/kurtosis-tech/parent/child: github.com/mieubrisse/new-child
```

...and you have Starlark code inside your package that looks like this...

```python
child = import_module("github.com/kurtosis-tech/parent/child/main.star")

 
def run(plan):
    child.run(plan)
```

...then Kurtosis would use the "more specific" rule to replace `github.com/kurtosis-tech/parent/child/main.star` with `github.com/mieubrisse/new-child`.
The `import_module` call would then be functionally equivalent to the following:

```python
child = import_module("github.com/mieubrisse/new-child/main.star")
step_
```

In other words, replace directives are matched “longest match first”.

<!----------------------- ONLY LINKS BELOW HERE ----------------------------->
[package]: ./packages.md
[how-do-kurtosis-imports-work-explanation]: ../advanced-concepts/how-do-kurtosis-imports-work.md
[locators]: ./locators.md
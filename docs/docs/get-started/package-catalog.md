---
title: Package Catalog
sidebar_label: Basic Concepts
slug: '/package-catalog'
sidebar_position: 6
---

Package catalog
---------------

The Kurtosis package catalog is a public repository of Kurtosis packages where package authors can publish their packages which later can be found and used by any user.

The Package Catalog can be found at: [catalog.kurtosis.com][package-catalog]


How can do I add my package to the catalog?
---------------------------------------------------

The catalog is made up of a curated list that is stored in the `kurtosis-package-catalog.yml` file within the [kurtosis-package-catalog repository.][package-catalog-repository]

If an author wants to add their package to the catalog, they should add its name on the list inside the `kurtosis-package-catalog.yml` file and send a pull request to be validated by the CI validations and approved by a Kurtosis administrator.

### What validates the CI?

The CI's automatic jobs validate the following for each new Pull Request created:

- That there are no duplicated names within the `kurtosis-package-catalog.yml` file, there cannot be two or more packages with the same name.
- That the package repository exists.
- That the repository contains the `kurtosis.yml` file.
- That the name declared in the `kurtosis.yml` file corresponds to the name that is being added to the catalog
- Only if the package contains an icon (you should upload an image file named `kurtosis-package-icon.png` in your package repository, at the same level as the `kurtosis.yml` file, is if you want to add an icon to be displayed in the catalog)
    - It will be validated that its size is greater than 120px
    - It will be validated that its size is less than 1024px
    - If it validated that it contains a 1:1 aspect ratio, a square image.

If the package does not pass these validations, you should make the necessary modifications until it is validated and then approved and added to the catalog.

[package-catalog]: https://catalog.kurtosis.com/
[package-catalog-repository]: https://github.com/kurtosis-tech/kurtosis-package-catalog

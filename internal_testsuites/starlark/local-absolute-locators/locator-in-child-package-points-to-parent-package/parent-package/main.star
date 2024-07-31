# This import should not fail because we're importing a file from another package
file = import_module("github.com/kurtosis-tech/child-package/file.star")


def run(plan):
    file.run(plan)
    pass

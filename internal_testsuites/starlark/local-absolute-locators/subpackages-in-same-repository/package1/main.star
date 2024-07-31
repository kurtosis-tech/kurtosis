# This import should not fail because we're importing a file from another package on same repository
file = import_module(
    "github.com/kurtosis-tech/subpackages-in-same-repository/package2/file.star"
)


def run(plan):
    file.run(plan)
    pass

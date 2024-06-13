# This import should fail because we're importing a file within the package
file = import_module(
    "github.com/kurtosis-tech/block-package-local-absolute-locator/file.star"
)


def run(plan):
    pass

# This import should be allowed because it's loaded in the child package
file_in_parent = import_module(
    "github.com/kurtosis-tech/parent-package/file-in-parent.star"
)


def run(plan):
    plan.print(file_in_parent.FOO)

lib = import_module("github.com/kurtosis-tech/kurtosis/internal_testsuites/starlark/valid-kurtosis-package-with-input/lib/lib.star")


def run(plan, args):
    plan.print(args["greetings"])
    output = struct(message="Hello " + lib.world + "!")
    plan.print(output.message)
    return output

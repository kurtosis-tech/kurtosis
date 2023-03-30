lib = import_module("github.com/sample/sample-kurtosis-package/lib/lib.star")


def run(plan, args):
    plan.print(args["greetings"])
    output = struct(message="Hello " + lib.world + "!")
    plan.print(output.message)
    return output

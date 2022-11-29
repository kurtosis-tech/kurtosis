lib = import_module("github.com/sample/sample-kurtosis-package/lib/lib.star")


def run(args):
    print(args.greetings)
    output = struct(message="Hello " + lib.world + "!")
    print(output.message)
    return output

lib = import_module("github.com/sample/sample-kurtosis-module/lib/lib.star")
types = import_types("github.com/sample/sample-kurtosis-module/types.proto")


def run(input_args):
    print(input_args.greetings)
    output = types.ModuleOutput({
        "message": "Hello " + lib.world + "!"
    })
    print(output.message)
    return output

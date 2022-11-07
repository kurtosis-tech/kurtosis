load("github.com/sample/sample-kurtosis-module/lib/lib.star", "world")
types = import_types("github.com/sample/sample-kurtosis-module/types.proto")


def main(input_args):
    print(input_args.greetings)
    output = types.ModuleOutput({
        "message": "Hello " + world + "!"
    })
    print(output.message)
    return output

types = import_types("github.com/sample/sample-kurtosis-module/types.proto")


def run(input_args):
    output = types.ModuleOutput({
        "message": "Hello world!"
    })
    print(output.message)
    return output

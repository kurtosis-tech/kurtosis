load("github.com/sample/sample-kurtosis-module/lib/lib.star", "world")
types = import_types("github.com/sample/sample-kurtosis-module/types.proto")


def main():
    output = types.ModuleOutput({
        "greetings": "Hello " + world + "!"
    })
    print(output.greetings)
    return output

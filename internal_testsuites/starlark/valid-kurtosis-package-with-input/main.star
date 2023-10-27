lib = import_module("./lib/lib.star")


def run(plan, args):
    plan.print(args["greetings"])
    output = struct(message="Hello " + lib.world + "!")
    plan.print(output.message)
    return output

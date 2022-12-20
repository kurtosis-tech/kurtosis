def run(plan):
    output = struct(message="Hello world!")
    plan.print(output.message)
    return output

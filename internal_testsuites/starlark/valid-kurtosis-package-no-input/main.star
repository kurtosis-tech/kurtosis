def run(plan):
    output = struct(message="package with no input")
    plan.print(output.message)
    return output

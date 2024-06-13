def run(plan, args):
    plan.verify(args["name"], "==", "John Doe")
    plan.print(args["name"])

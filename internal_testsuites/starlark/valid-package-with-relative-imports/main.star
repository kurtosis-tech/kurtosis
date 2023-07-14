lib = import_module("./src/lib.star")

def run(plan):
    plan.print(lib.NAME)
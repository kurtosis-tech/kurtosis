lib = import_module("./src/lib.star")
password = read_file("./static_files/password.txt")

def run(plan):
    plan.print(lib.NAME)
    plan.print(password)
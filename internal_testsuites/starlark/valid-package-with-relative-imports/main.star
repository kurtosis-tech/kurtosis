lib = import_module("./src/lib.star")
password = read_file("./static_files/password.txt")


def run(plan):
    plan.upload_files("./static_files/password.txt", "upload")
    plan.print(lib.NAME)
    plan.print(password)

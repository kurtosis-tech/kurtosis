dependency = import_module("github.com/kurtosis-tech/sample-dependency-package/main.star")

EXPECTED_MSG_FROM_MAIN = "dependency-loaded-from-main"
EXPECTED_MSG_FROM_BRANCH = "dependency-loaded-from-test-branch"

MSG_ORIGIN_MAIN = "main"
MSG_ORIGIN_BRANCH = "branch"

# TODO remove https://github.com/kurtosis-tech/sample-startosis-load/tree/main/sample-package if it's not used
def run(plan, message_origin=MSG_ORIGIN_MAIN):
    plan.print("Replace with no main branch package loaded.")

    msg_from_dependency = dependency.get_msg()

    if message_origin == MSG_ORIGIN_MAIN:
        expected_msg = EXPECTED_MSG_FROM_MAIN
    elif message_origin == MSG_ORIGIN_BRANCH:
        expected_msg = EXPECTED_MSG_FROM_BRANCH
    else:
        expected_msg = ""

    plan.verify(
        value = expected_msg,
        assertion = "==",
        target_value = msg_from_dependency,
    )

    return

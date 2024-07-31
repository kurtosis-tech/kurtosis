dependency = import_module(
    "github.com/kurtosis-tech/sample-dependency-package/main.star"
)

EXPECTED_MSG_FROM_MAIN = "dependency-loaded-from-main"

MSG_ORIGIN_MAIN = "main"


def run(plan, message_origin=MSG_ORIGIN_MAIN):
    plan.print("Without replace package loaded.")

    msg_from_dependency = dependency.get_msg()

    if message_origin == MSG_ORIGIN_MAIN:
        expected_msg = EXPECTED_MSG_FROM_MAIN
    else:
        expected_msg = ""

    plan.verify(
        value=expected_msg,
        assertion="==",
        target_value=msg_from_dependency,
    )

    return

dependency = import_module(
    "github.com/kurtosis-tech/sample-dependency-package/main.star"
)

EXPECTED_MSG_FROM_MAIN = "dependency-loaded-from-main"
EXPECTED_MSG_FROM_LOCAL_PACKAGE_MAIN = "msg-loaded-from-local-dependency"

MSG_ORIGIN_MAIN = "main"
MSG_ORIGIN_LOCAL_DEPENDENCY = "local"


def run(plan, message_origin=MSG_ORIGIN_MAIN):
    plan.print("Replace with local package loaded.")

    msg_from_dependency = dependency.get_msg()

    if message_origin == MSG_ORIGIN_MAIN:
        expected_msg = EXPECTED_MSG_FROM_MAIN
    elif message_origin == MSG_ORIGIN_LOCAL_DEPENDENCY:
        expected_msg = EXPECTED_MSG_FROM_LOCAL_PACKAGE_MAIN
    else:
        expected_msg = ""

    plan.verify(
        value=expected_msg,
        assertion="==",
        target_value=msg_from_dependency,
    )

    return

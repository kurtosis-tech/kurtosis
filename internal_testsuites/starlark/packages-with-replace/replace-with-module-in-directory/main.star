internal_dependency = import_module(
    "github.com/kurtosis-tech/sample-dependency-package/directory/internal-module.star"
)

EXPECTED_MSG_FROM_INTERNAL_MODULE_MAIN = (
    "dependency-loaded-from-internal-module-in-main-branch"
)
EXPECTED_MSG_FROM_ANOTHER_PACKAGE_INTERNAL_MODULE_MAIN = (
    "another-dependency-loaded-from-internal-module-in-main-branch"
)

MSG_ORIGIN_FROM_SAMPLE = "sample"
MSG_ORIGIN_FROM_ANOTHER_SAMPLE = "another-sample"


def run(plan, message_origin=MSG_ORIGIN_FROM_SAMPLE):
    plan.print("Replace with module in directory sample package loaded.")

    msg_from_dependency = internal_dependency.get_msg()

    if message_origin == MSG_ORIGIN_FROM_SAMPLE:
        expected_msg = EXPECTED_MSG_FROM_INTERNAL_MODULE_MAIN
    elif message_origin == MSG_ORIGIN_FROM_ANOTHER_SAMPLE:
        expected_msg = EXPECTED_MSG_FROM_ANOTHER_PACKAGE_INTERNAL_MODULE_MAIN
    else:
        expected_msg = ""

    plan.verify(
        value=expected_msg,
        assertion="==",
        target_value=msg_from_dependency,
    )

    return

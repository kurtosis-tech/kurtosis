MSG = "msg-loaded-from-local-dependency"


def run(plan, args):
    plan.print("Local sample dependency package loaded.")

    msg_to_return = get_msg()

    return msg_to_return


def get_msg():
    return MSG

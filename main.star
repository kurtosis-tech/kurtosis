ethereum = import_module("github.com/kurtosis-tech/ethereum-package/main.star")

def run(plan, args):
    plan.run_sh(
            run = "mkdir -p kurtosis && echo $(ls)",
    )

    ethereum.run(plan)
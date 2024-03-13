def run(plan, args):
    plan.add_service(
        name="service",
        config=ServiceConfig(
            image=ImageBuildSpec(
                image_name="kurtosistech/service",
                build_context_dir="./",
                build_args={
                    "BUILD_ARG": "VALUE",
                }
            ),
        )
    )

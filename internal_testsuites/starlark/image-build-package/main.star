def run(plan, args):
    plan.add_service(
        name="service-1",
        config=ServiceConfig(
            image=ImageBuildSpec(
                image_name="kurtosistech/service",
                build_context_dir="./",
                build_args={
                    "BUILD_ARG": "VALUE",
                },
            ),
        ),
    )

    plan.add_service(
        name="service-2",
        config=ServiceConfig(
            image=ImageBuildSpec(
                image_name="kurtosistech/service",
                build_context_dir="./test",
                build_file="test.Dockerfile",
            )
        ),
    )

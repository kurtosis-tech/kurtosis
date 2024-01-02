def run(plan):
    plan.add_service(
        name = "this_is_not_valid",
        config = ServiceConfig(
            image = "postgres:15.2-alpine"
        )
    )
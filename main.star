def run(plan, args):
    config = ServiceConfig(
                image="alpine"
            )

    service = plan.add_service(
        name="lol",
        config=config,
    )

    plan.exec(
        service_name="lol",
		recipe = ExecRecipe(
			command = ["wget", "-O", "./output.txt", "https://raw.githubusercontent.com/ethereum/c-kzg-4844/main/src/trusted_setup.txt"]
		)
    )

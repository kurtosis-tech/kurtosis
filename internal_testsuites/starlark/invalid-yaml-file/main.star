
def run(plan, args):
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
	)

	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		},
        ready_conditions = ready_conditions
	)

	service = plan.add_service(name = "test", config = service_config)

    plan.wait(
        service_name = service.name, 
        recipe = get_recipe, 
        field = "code", 
        assertion = "==", 
        target_value = 200, 
        interval = "1s", 
        timeout = "40s"
    )

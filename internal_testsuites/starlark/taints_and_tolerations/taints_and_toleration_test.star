def run(plan, with_tolerations=True):
    if with_toleration:
        config = ServiceConfig(
            image="kurtosistech/example-datastore-server",
            tolerations=[
                Toleration(
                    key="foo", value="bar", operator="Equal", effect="NoSchedule"
                )
            ],
            ports={"grpc": PortSpec(number=1323, transport_protocol="TCP")},
        )
    else:
        config = ServiceConfig(
            image="kurtosistech/example-datastore-server",
            ports={"grpc": PortSpec(number=1323, transport_protocol="TCP")},
        )

    plan.add_service(
        name="test-data-store",
        config=config,
    )

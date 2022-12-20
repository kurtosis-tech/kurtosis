DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "example-datastore-server-startosis"
DATASTORE_PORT_ID = "grpc"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

def run(plan, args):
    plan.print("Adding service " + DATASTORE_SERVICE_ID + ".")

    config = ServiceConfig(
        image = DATASTORE_IMAGE,
        ports = {
            DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
        }
    )

    plan.add_service(service_id = DATASTORE_SERVICE_ID, config = config)
    plan.print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")

DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "example-datastore-server-startosis"
DATASTORE_PORT_ID = "grpc"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

config = struct(
    image = DATASTORE_IMAGE,
    ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    }
)

add_service(service_id = DATASTORE_SERVICE_ID, config = config)
print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")

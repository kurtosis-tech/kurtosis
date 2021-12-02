import {
  ServiceID,
  EnclaveContext,
  SharedPath,
  ContainerConfig,
  ContainerConfigBuilder,
  PortBinding,
  ServiceContext
} from "kurtosis-core-api-lib";
import { EngineServiceClient } from "kurtosis-engine-api-lib";
import * as datastoreApi from "example-datastore-server-api-lib";
import * as serverApi from "example-api-server-api-lib";
import { err, ok, Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "grpc";
import log from "loglevel";

const DATASTORE_IMAGE = "kurtosistech/example-datastore-server";

const DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS = 10;
const DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS = 1000;

const DATASTORE_PORT_STR = `${datastoreApi.LISTEN_PORT}/${datastoreApi.LISTEN_PROTOCOL}`;

export async function addDatastoreService(
  serviceId: ServiceID,
  enclaveCtx: EnclaveContext
): Promise<Result<
  {
    serviceCtx: ServiceContext;
    client: datastoreApi.DatastoreServiceClient;
    clientCloseFunc: () => void;
  },
  Error>> {

  const containerConfigSupplier = getDatastoreContainerConfigSupplier();

  const addServiceResult = await enclaveCtx.addService(serviceId, containerConfigSupplier);

  if (addServiceResult.isErr()) {
    return err(new Error("An error occurred adding the datastore service"));
  }

  const [serviceCtx, hostPortBindings] = addServiceResult.value;

  const hostPortBinding: PortBinding | undefined = hostPortBindings.get(DATASTORE_PORT_STR);

  if (hostPortBinding === undefined) {
    return err(new Error(`No datastore host port binding found for port string ${DATASTORE_PORT_STR}`));
  }

  const datastoreIp = hostPortBinding.getInterfaceIp();
  const datastorePortNumStr = hostPortBinding.getInterfacePort();

  const createDatastoreClientResult = createDatastoreClient(
    datastoreIp,
    datastorePortNumStr
  );

  if (createDatastoreClientResult.isErr()) {
    log.info(`An error occurred creating the datastore client for IP ${datastoreIp} and port ${datastorePortNumStr}`)
    return err(createDatastoreClientResult.error);
  }

  const { client, clientCloseFunc } = createDatastoreClientResult.value;

  const waitForHealthyResult = await waitForHealthy(
    client,
    DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS,
    DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS
  );

  if (waitForHealthyResult.isErr()) {
    log.info("An error occurred waiting for the datastore service to become available")
    return err(waitForHealthyResult.error);
  }

  return ok({ serviceCtx, client, clientCloseFunc });
}

type CreateDatastoreClientReturn = Result<{ client: datastoreApi.DatastoreServiceClient; clientCloseFunc: () => void }, Error>;

function createDatastoreClient(ipAddr: string, portNum: string): CreateDatastoreClientReturn {

  const url = `${ipAddr}:${portNum}`;

  let conn: Result<EngineServiceClient, Error>;

  try {
    conn = ok(new EngineServiceClient(url, grpc.credentials.createInsecure()));
  } catch (error) {
    log.error(`An error occurred connecting to datastore service on URL ${url}`);
    return err(new Error());
  }

  const clientCloseFunc = () => {
    if (conn.isOk()) conn.value.close();
  };

  const client = new datastoreApi.DatastoreServiceClient(url, grpc.credentials.createInsecure());

  return ok({
    client,
    clientCloseFunc,
  });
}

async function waitForHealthy(
  client: datastoreApi.DatastoreServiceClient | serverApi.ExampleAPIServerServiceClient,
  retries: number,
  retriesDelayMilliseconds: number
): Promise<Result<null, Error>> {

  const emptyArgs: google_protobuf_empty_pb.Empty =
    new google_protobuf_empty_pb.Empty();

  const sleep = (ms: number) =>
    new Promise((resolve) => setTimeout(resolve, ms));

  const checkClientAvailabilityPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise(async (resolve, _unusedReject) => {
    client.isAvailable(emptyArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
      if (error === null) {
        if (!response) {
          resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
        } else {
          resolve(ok(response));
        }
      } else {
        resolve(err(error));
      }
    }
    );
  });

  for (let i = 0; i < retries; i++) {
    const result = await checkClientAvailabilityPromise;
    if (result.isOk()) return ok(null);

    await sleep(retriesDelayMilliseconds);
  }

  return err(new Error(`The service didn't return a success code, even after ${retries} retries with ${retriesDelayMilliseconds} milliseconds in between retries`));

}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================

function getDatastoreContainerConfigSupplier(): (
  ipAddr: string,
  sharedDirectory: SharedPath
) => Result<ContainerConfig, Error> {

  const containerConfigSupplier = (
    ipAddr: string,
    sharedDirectory: SharedPath
  ): Result<ContainerConfig, Error> => {

    const datastorePortsSet = new Set().add(DATASTORE_PORT_STR) as Set<string>;

    const containerConfig = new ContainerConfigBuilder(DATASTORE_IMAGE).withUsedPorts(datastorePortsSet).build();

    return ok(containerConfig);
  };

  return containerConfigSupplier;
}
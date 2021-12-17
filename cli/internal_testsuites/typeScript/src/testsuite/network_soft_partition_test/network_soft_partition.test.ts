import { 
    ContainerConfig, 
    ContainerConfigBuilder, 
    EnclaveContext, 
    PartitionID, 
    PortProtocol, 
    PortSpec, 
    ServiceID, 
    SharedPath, 
    SoftPartitionConnection,
    UnblockedPartitionConnection
} from "kurtosis-core-api-lib";
import { PartitionConnection } from "kurtosis-core-api-lib/build/lib/enclaves/partition_connection";
import log from "loglevel";
import { ok, Result, err } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "network-soft-partition-test";
const IS_PARTITIONING_ENABLED = true;

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started";
const EXAMPLE_SERVICE_ID: ServiceID = "docker-getting-started";
const KURTOSIS_IP_ROUTE_2_DOCKER_IMAGE_NAME = "kurtosistech/iproute2";
const TEST_SERVICE: ServiceID = "test-service";
const EXAMPLE_SERVICE_PORT_NUM_INSIDE_NETWORK = 80;

const EXEC_COMMAND_SUCCESS_EXIT_CODE = 0;

const EXAMPLE_SERVICE_PARTITION_ID: PartitionID = "example";
const TEST_SERVICE_PARTITION_ID: PartitionID = "test";

const EXAMPLE_SERVICE_MAIN_PORT_ID = "main-port";

const SLEEP_CMD = "sleep";

const TEST_SERVICE_SLEEP_MILLISECONDS_STR = "300000";

const PERCENTAGE_SIGN = "%";
const ZERO_PACKET_LOSS = 0;
const SOFT_PARTITION_PACKET_LOSS_PERCENTAGE = 40;

const HUNDRED_SECONDS_IN_MILLISECONDS = 100000
jest.setTimeout(HUNDRED_SECONDS_IN_MILLISECONDS);

interface MtrReport {
    report: {
        hubs: Array<{"Loss%": number}>
    }
}

test("Test network soft partitions", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const configSupplier = getExampleServiceConfigSupplier()

        const exampleAddServiceResult = await enclaveContext.addService(EXAMPLE_SERVICE_ID, configSupplier)
        if(exampleAddServiceResult.isErr()){
            log.error("An error occurred adding the datastore service")
            throw exampleAddServiceResult.error
        }

        const exampleServiceContext = exampleAddServiceResult.value
        log.debug(`Example service IP: ${exampleServiceContext.getPrivateIPAddress()}`)


        const containerConfigSupplier = getTestServiceContainerConfigSupplier()

        const testAddServiceResult = await enclaveContext.addService(TEST_SERVICE, containerConfigSupplier)
        if(testAddServiceResult.isErr()){
            log.error("An error occurred adding the file server service")
            throw testAddServiceResult.error
        }

        const testServiceContext = testAddServiceResult.value
        log.debug(`Test service IP: ${testServiceContext.getPrivateIPAddress()}`)

        const installMtrCmd = ["apk", "add", "mtr"]

        {
            const execCommandResult = await testServiceContext.execCommand(installMtrCmd)
            if(execCommandResult.isErr()){
                log.error(`An error occurred executing command "${installMtrCmd}" `)
                throw execCommandResult.error
            }
            
            const [exitCode] = execCommandResult.value
            
            if(EXEC_COMMAND_SUCCESS_EXIT_CODE !== exitCode) {
                throw new Error(`Command "${installMtrCmd}" to install mtr cli exited with non-successful exit code "${exitCode}"`)
            }
        }

        // ------------------------------------- TEST RUN ----------------------------------------------
        log.info("Executing mtr report to check there is no packet loss in services' communication before soft partition...")

        const mtrReportCmd = [
            "mtr", 
            exampleServiceContext.getPrivateIPAddress(),
            "--report",
            "--json",
            "--report-cycles",
            "4", // We set report cycles to 4 to generate the report faster because default is 10
            "--no-dns", //No domain name resolution, also to improve velocity
        ];

        {
            const execCommandResult = await testServiceContext.execCommand(mtrReportCmd)
            if(execCommandResult.isErr()){
                log.error(`An error occurred executing command "${mtrReportCmd}"`)
                throw execCommandResult.error
            }
            
            const [ exitCode, logOutput ] = execCommandResult.value
            
            if(EXEC_COMMAND_SUCCESS_EXIT_CODE !== exitCode) {
                throw new Error(`Command "${mtrReportCmd}" to run mtr report exited with non-successful exit code "${exitCode}"`)
            }
            
            const jsonStr = logOutput
            log.debug(`MTR report before soft partition result:\n ${jsonStr}`)
            
            let mtrReportBeforeSoftPartition: MtrReport;
            
            try{
                mtrReportBeforeSoftPartition = JSON.parse(jsonStr) 
            }catch(error){
                log.error(`An error occurred unmarshalling json string "${jsonStr}" to mtr report struct`)
                if(error instanceof Error){
                    throw error
                }else{
                    throw new Error("Encountered error while writing the file, but the error wasn't of type Error")
                }
            }
            
            const packetLoss = mtrReportBeforeSoftPartition.report.hubs[0]["Loss%"]
            if(ZERO_PACKET_LOSS !== packetLoss){
                throw new Error(`Expected zero packet loss before soft partitioning, but packet loss was ${packetLoss}`)
            }
        }

        log.info("Report complete successfully, there was no packet loss between services during the test")

        log.info(`Executing soft partition with packet loss ${SOFT_PARTITION_PACKET_LOSS_PERCENTAGE}${PERCENTAGE_SIGN}...`)
        
        const softPartitionConnection = new SoftPartitionConnection(SOFT_PARTITION_PACKET_LOSS_PERCENTAGE)
        
        const repartitionNetworkResult = await repartitionNetwork(enclaveContext, softPartitionConnection)

        if(repartitionNetworkResult.isErr()){
            log.error("An error occurred executing repartition network")
            throw repartitionNetworkResult.error
        }

        log.info("Partition complete")

        log.info("Executing mtr report to check there is packet loss in services' communication after soft partition...")

        {
            const execCommandResult = await testServiceContext.execCommand(mtrReportCmd)
            
            if(execCommandResult.isErr()){
                log.error(`An error occurred executing command '${mtrReportCmd}'`)
                throw execCommandResult.error
            }
            
            const [ exitCode, logOutput ] = execCommandResult.value
            
            if(EXEC_COMMAND_SUCCESS_EXIT_CODE !== exitCode){
                throw new Error(`Command "${mtrReportCmd}" to run mtr report exited with non-successful exit code "${EXEC_COMMAND_SUCCESS_EXIT_CODE}"`)
            }
            
            const jsonStr = logOutput
            log.debug(`MTR report after soft partition result:\n  newJsonStr`)
            
            let mtrReportAfterPartition: MtrReport;
            
            try{
                mtrReportAfterPartition = JSON.parse(jsonStr) 
            }catch(error){
                log.error(`An error occurred unmarshalling json string "${jsonStr}" to mtr report struct`)
                if(error instanceof Error){
                    throw error
                }else{
                    throw new Error("Encountered error while writing the file, but the error wasn't of type Error")
                }
            }

            const packetLoss = mtrReportAfterPartition.report.hubs[0]["Loss%"]
            if(ZERO_PACKET_LOSS === packetLoss){
                throw new Error(`Expected nonzero packet loss after applying the soft partition, but packet loss ${packetLoss}`)
            }
            
            log.info(`Report complete successfully, there was ${packetLoss}${PERCENTAGE_SIGN} packet loss between services during the test`)
        }

        log.info("Executing repartition network to unblock partition and join services again...")

        const unblockedPartitionConnection = new UnblockedPartitionConnection()

        {
            const repartitionNetworkResult = await repartitionNetwork(enclaveContext, unblockedPartitionConnection)
            
            if(repartitionNetworkResult.isErr()){
                log.error("An error occurred executing repartition network")
                throw repartitionNetworkResult.error
            }
        }
        
        log.info("Partitions unblocked successfully")

        log.info("Executing mtr report to check there is no packet loss in services' communication after unblocking partition...")

        {
            const execCommandResult = await testServiceContext.execCommand(mtrReportCmd)
            if(execCommandResult.isErr()){
                log.error(`An error occurred executing command '${mtrReportCmd}' `)
                throw execCommandResult.error
            }

            const [exitCode, logOutput] = execCommandResult.value
            if(EXEC_COMMAND_SUCCESS_EXIT_CODE !== exitCode){
                throw new Error(`Command '${mtrReportCmd}' to run mtr report exited with non-successful exit code '${exitCode}'`)
            }
            
            const jsonStr = logOutput
            log.debug(`MTR report after unblocking partition result:\n  ${jsonStr}`)

            let mtrReportAfterUnblockedPartition: MtrReport
            
            try{
                mtrReportAfterUnblockedPartition = JSON.parse(jsonStr)
            }catch(error){
                log.error(`An error occurred unmarshalling json string "${jsonStr}" to mtr report struct`)
                if(error instanceof Error){
                    throw error
                }else{
                    throw new Error("Encountered error while writing the file, but the error wasn't of type Error")
                }
            }

            const packetLoss = mtrReportAfterUnblockedPartition.report.hubs[0]["Loss%"]
            if(ZERO_PACKET_LOSS !== packetLoss){
                throw new Error(`Expected zero packet loss after removing the soft partition, but packet loss was ${packetLoss}`)
            }

	        log.info("Report complete successfully, there was no packet loss between services during the test")
        }

    }finally{
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})

function getExampleServiceConfigSupplier():(ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error>{
    const portSpec = new PortSpec(EXAMPLE_SERVICE_PORT_NUM_INSIDE_NETWORK, PortProtocol.TCP);
    const containerConfigSupplier = (ipAddr: string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {
        const exampleServiceMainPortId = new Map<string,PortSpec>()
        exampleServiceMainPortId.set(EXAMPLE_SERVICE_MAIN_PORT_ID,portSpec)
        const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
                .withUsedPorts(exampleServiceMainPortId)
                .build()
        return ok(containerConfig)
    }

    return containerConfigSupplier
}

function getTestServiceContainerConfigSupplier():(ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {
    const containerConfigSupplier = (ipAddr: string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {
        
        // We sleep because the only function of this container is to test Docker executing a command while it's running
		// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
		// args), but this provides a nice little regression test of the ENTRYPOINT overriding
        const entrypointArgs = [SLEEP_CMD]
        const cmdArgs = [TEST_SERVICE_SLEEP_MILLISECONDS_STR]
        
        const containerConfig = new ContainerConfigBuilder(KURTOSIS_IP_ROUTE_2_DOCKER_IMAGE_NAME)
                .withEntrypointOverride(entrypointArgs)
                .withCmdOverride(cmdArgs)
                .build()
        
        return ok(containerConfig)
    }

    return containerConfigSupplier
}

async function repartitionNetwork(enclaveContext: EnclaveContext, partitionConnection: PartitionConnection): Promise<Result<null, Error>> {
    const partitionServices = new Map<PartitionID,Set<ServiceID>>()
    partitionServices.set(EXAMPLE_SERVICE_PARTITION_ID, new Set([EXAMPLE_SERVICE_ID]))
    partitionServices.set(TEST_SERVICE_PARTITION_ID, new Set([TEST_SERVICE]))

    const partitionConnections = new Map<PartitionID, Map<PartitionID,PartitionConnection>>()
    partitionConnections.set(EXAMPLE_SERVICE_PARTITION_ID, new Map([[TEST_SERVICE_PARTITION_ID, partitionConnection]]))

    const defaultPartitionConnection = partitionConnection

    const repartitionNetworkResult = await enclaveContext.repartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection)

    if(repartitionNetworkResult.isErr()){
        log.error(`An error occurred repartitioning the network with partition connection = ${partitionConnection}`)
        return err(repartitionNetworkResult.error)
    }

    return ok(null)
}
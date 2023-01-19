import {ServiceUUID, ServiceLog, ServiceLogsStreamContent} from "kurtosis-sdk";
import {Readable} from "stream";

export class ReceivedStreamContent {
    readonly receivedLogLinesByService: Map<ServiceUUID, Array<ServiceLog>>;
    readonly receivedNotFoundServiceUuids: Set<ServiceUUID>;

    constructor(
        receivedLogLinesByService: Map<ServiceUUID, Array<ServiceLog>>,
        receivedNotFoundServiceUuids: Set<ServiceUUID>,
    ) {
        this.receivedLogLinesByService = receivedLogLinesByService;
        this.receivedNotFoundServiceUuids = receivedNotFoundServiceUuids;
    }
}

export function receiveExpectedLogLinesFromServiceLogsReadable(
    serviceLogsReadable: Readable,
    expectedLogLinesByService: Map<ServiceUUID, ServiceLog[]>,
): Promise<ReceivedStreamContent> {
    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = new Promise<ReceivedStreamContent>((resolve, _unusedReject) => {

        let receivedLogLinesByService: Map<ServiceUUID, Array<ServiceLog>> = new Map<ServiceUUID, Array<ServiceLog>>;
        let receivedNotFoundServiceUuids: Set<ServiceUUID> = new Set<ServiceUUID>;

        let allExpectedLogLinesWhereReceived = false;

        serviceLogsReadable.on('data', (serviceLogsStreamContent: ServiceLogsStreamContent) => {
            const serviceLogsByUuid: Map<ServiceUUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceUuids();
            receivedNotFoundServiceUuids = serviceLogsStreamContent.getNotFoundServiceUuids();

            for (let [serviceUuid, serviceLogLines] of serviceLogsByUuid) {
                let receivedLogLines: ServiceLog[] = new Array<ServiceLog>;
                if(receivedLogLinesByService.has(serviceUuid)){
                    const userServiceLogLines: ServiceLog[] | undefined = receivedLogLinesByService.get(serviceUuid)
                    if (userServiceLogLines !== undefined) {
                        receivedLogLines = userServiceLogLines.concat(serviceLogLines)
                    }
                } else {
                    receivedLogLines = serviceLogLines;
                }
                receivedLogLinesByService.set(serviceUuid, receivedLogLines)
            }

            for (let [serviceUuid, expectedLogLines] of expectedLogLinesByService) {
                if (expectedLogLines === undefined && !receivedLogLinesByService.has(serviceUuid)) {
                    break;
                }

                if (expectedLogLines.length < 0 && !receivedLogLinesByService.has(serviceUuid)) {
                    break;
                }

                let receivedLogLines: ServiceLog[] | undefined = receivedLogLinesByService.get(serviceUuid);

                if (receivedLogLines === undefined) {
                    receivedLogLines = new Array<ServiceLog>;
                }

                if (expectedLogLines.length !== receivedLogLines.length) {
                    break;
                }
                allExpectedLogLinesWhereReceived = true
            }
            if (allExpectedLogLinesWhereReceived) {
                serviceLogsReadable.destroy()
                const receivedStreamContent: ReceivedStreamContent = new ReceivedStreamContent(
                    receivedLogLinesByService,
                    receivedNotFoundServiceUuids,
                )
                resolve(receivedStreamContent)
            }
        })

        serviceLogsReadable.on('error', function (readableErr: { message: any; }) {
            if (!serviceLogsReadable.destroyed) {
                serviceLogsReadable.destroy()
                throw new Error(`Expected read all user service logs but an error was received from the user service readable object.\n Error: "${readableErr.message}"`)
            }
        })

        serviceLogsReadable.on('end', function () {
            if (!serviceLogsReadable.destroyed) {
                serviceLogsReadable.destroy()
                throw new Error("Expected read all user service logs but the user service readable logs object has finished before reading all the logs.")
            }
        })

    })

    return receivedStreamContentPromise;
}

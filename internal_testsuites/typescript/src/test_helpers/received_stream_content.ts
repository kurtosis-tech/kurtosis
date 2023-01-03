import {ServiceGUID, ServiceLog, ServiceLogsStreamContent} from "kurtosis-sdk";
import {Readable} from "stream";

export class ReceivedStreamContent {
    readonly receivedLogLinesByService: Map<ServiceGUID, Array<ServiceLog>>;
    readonly receivedNotFoundServiceGuids: Set<ServiceGUID>;

    constructor(
        receivedLogLinesByService: Map<ServiceGUID, Array<ServiceLog>>,
        receivedNotFoundServiceGuids: Set<ServiceGUID>,
    ) {
        this.receivedLogLinesByService = receivedLogLinesByService;
        this.receivedNotFoundServiceGuids = receivedNotFoundServiceGuids;
    }
}

export function receiveExpectedLogLinesFromServiceLogsReadable(
    serviceLogsReadable: Readable,
    expectedLogLinesByService: Map<ServiceGUID, ServiceLog[]>,
): Promise<ReceivedStreamContent> {
    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = new Promise<ReceivedStreamContent>((resolve, _unusedReject) => {

        let receivedLogLinesByService: Map<ServiceGUID, Array<ServiceLog>> = new Map<ServiceGUID, Array<ServiceLog>>;
        let receivedNotFoundServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>;

        let allExpectedLogLinesWhereReceived = false;

        serviceLogsReadable.on('data', (serviceLogsStreamContent: ServiceLogsStreamContent) => {
            const serviceLogsByGuid: Map<ServiceGUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceGuids();
            receivedNotFoundServiceGuids = serviceLogsStreamContent.getNotFoundServiceGuids();

            for (let [serviceGuid, serviceLogLines] of serviceLogsByGuid) {
                let receivedLogLines: ServiceLog[] = new Array<ServiceLog>;
                if(receivedLogLinesByService.has(serviceGuid)){
                    const userServiceLogLines: ServiceLog[] | undefined = receivedLogLinesByService.get(serviceGuid)
                    if (userServiceLogLines !== undefined) {
                        receivedLogLines = userServiceLogLines.concat(serviceLogLines)
                    }
                } else {
                    receivedLogLines = serviceLogLines;
                }
                receivedLogLinesByService.set(serviceGuid, receivedLogLines)
            }

            for (let [serviceGuid, expectedLogLines] of expectedLogLinesByService) {
                if (expectedLogLines === undefined && !receivedLogLinesByService.has(serviceGuid)) {
                    break;
                }

                if (expectedLogLines.length < 0 && !receivedLogLinesByService.has(serviceGuid)) {
                    break;
                }

                let receivedLogLines: ServiceLog[] | undefined = receivedLogLinesByService.get(serviceGuid);

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
                    receivedNotFoundServiceGuids,
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

import {ServiceGUID, ServiceLog, ServiceLogsStreamContent} from "kurtosis-sdk";
import {Readable} from "stream";

export class ReceivedStreamContent {
    readonly receivedLogLines: Array<ServiceLog>;
    readonly receivedNotFoundServiceGuids: Set<ServiceGUID>;

    constructor(
        receivedLogLines: Array<ServiceLog>,
        receivedNotFoundServiceGuids: Set<ServiceGUID>,
    ) {
        this.receivedLogLines = receivedLogLines;
        this.receivedNotFoundServiceGuids = receivedNotFoundServiceGuids;
    }
}

export function newReceivedStreamContentPromise(
    serviceLogsReadable: Readable,
    serviceGuid: string,
    expectedLogLines: string[],
    expectedNonExistenceServiceGuids: Set<ServiceGUID>,
): Promise<ReceivedStreamContent> {
    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = new Promise<ReceivedStreamContent>((resolve, _unusedReject) => {

        serviceLogsReadable.on('data', (serviceLogsStreamContent: ServiceLogsStreamContent) => {
            const serviceLogsByServiceGuids: Map<ServiceGUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceGuids()
            const notFoundServiceGuids: Set<ServiceGUID> = serviceLogsStreamContent.getNotFoundServiceGuids()

            let receivedLogLines: Array<ServiceLog> | undefined = serviceLogsByServiceGuids.get(serviceGuid);

            if (expectedNonExistenceServiceGuids.size > 0) {
                if(receivedLogLines !== undefined){
                    throw new Error(`Expected to receive undefined log lines but these log lines content ${receivedLogLines} was received instead`)
                }
            } else {
                if(receivedLogLines === undefined){
                    throw new Error("Expected to receive log lines content but and undefined value was received instead")
                }
            }

            if(receivedLogLines === undefined) {
                receivedLogLines = new Array<ServiceLog>()
            }

            if (receivedLogLines.length === expectedLogLines.length) {
                const receivedStreamContent: ReceivedStreamContent = new ReceivedStreamContent(
                    receivedLogLines,
                    notFoundServiceGuids,
                )
                serviceLogsReadable.destroy()
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

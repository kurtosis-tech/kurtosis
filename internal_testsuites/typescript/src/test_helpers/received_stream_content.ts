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

export function newReceivedStreamContentPromise(
    serviceLogsReadable: Readable,
): Promise<ReceivedStreamContent> {
    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = new Promise<ReceivedStreamContent>((resolve, _unusedReject) => {

        serviceLogsReadable.on('data', (serviceLogsStreamContent: ServiceLogsStreamContent) => {
            const receivedLogLinesByService: Map<ServiceGUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceGuids();
            const notFoundServiceGuids: Set<ServiceGUID> = serviceLogsStreamContent.getNotFoundServiceGuids();

            const receivedStreamContent: ReceivedStreamContent = new ReceivedStreamContent(
                receivedLogLinesByService,
                notFoundServiceGuids,
            )
            serviceLogsReadable.destroy()
            resolve(receivedStreamContent)

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

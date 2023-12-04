import { useState } from "react";import streamsaver from "streamsaver";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../emui/enclaves/types";
import { DownloadButton } from "./DownloadButton";
import {ServiceInfo, StreamedDataChunk} from "enclave-manager-sdk/build/api_container_service_pb";
import {isDefined, stripAnsi} from "../utils";
import { Timestamp } from "@bufbuild/protobuf";
import {LogLineMessage} from "./enclaves/logs/types";
import {DateTime} from "luxon";
import {Simulate} from "react-dom/test-utils";
import canPlayThrough = Simulate.canPlayThrough;

type DownloadLogsButtonProps = {
    logsFileName:string,
    enclave:EnclaveFullInfo,
    service?: ServiceInfo,
    logsToDownload: LogLineMessage[];
};

export const DownloadLogsButton = ({ logsFileName, enclave, service, logsToDownload }: DownloadLogsButtonProps) => {
    const kurtosisClient = useKurtosisClient();
    const [isLoading, setIsLoading] = useState(false);
    // const [logLinesToDownload, setLogLinesToDownload] = useState(logsToDownload);

    const serviceLogLineToLogLineMessage = (lines: string[], timestamp?: Timestamp): LogLineMessage[] => {
        return lines.map((line) => ({
            message: line,
            timestamp: isDefined(timestamp) ? DateTime.fromJSDate(timestamp?.toDate()) : undefined,
        }));
    };

    const handleDownloadLogsClick = async () => {
        setIsLoading(true);
        const writableStream = streamsaver.createWriteStream(logsFileName || "logs.txt");
        const writer = writableStream.getWriter();

        if (service){
            console.log("pulling logs")
            const abortController = new AbortController();
            for await (const lineGroup of await kurtosisClient.getServiceLogs(abortController, enclave, [service], false, 0, true)) {
                try {
                    console.log(lineGroup)
                    const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
                    if (!isDefined(lineGroupForService)) {
                        console.error("smth went wrong")
                        continue;
                    }
                    const parsedLogLines = serviceLogLineToLogLineMessage(lineGroupForService.line, lineGroupForService.timestamp);
                    console.log("downloading logs")
                    const logBlob = new Blob([parsedLogLines.map(({message}) => message)
                        .filter(isDefined)
                        .map(stripAnsi)
                        .join("\n")])
                    await writer.write(logBlob);
                } catch(err){
                    console.error(err)
                }
            }
        } else if (logsToDownload){
            console.log("downloading logs")
            await writer.write(getLogsValue())
        }

        await writer.close();
        console.log("finished downloading logs")
        setIsLoading(false);
    };

    const getLogsValue = () => {
        return logsToDownload
            .map(({ message }) => message)
            .filter(isDefined)
            .map(stripAnsi)
            .join("\n");
    };

    return (
        <DownloadButton
            size={"sm"}
            fileName={logsFileName || `logs.txt`}
            // isDisabled={logsToDownload.length === 0}
            isIconButton
            aria-label={"Download logs"}
            color={"gray.100"}
            isLoading={isLoading}
            onClick={handleDownloadLogsClick}
        />
    );
};

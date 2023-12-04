import { useState } from "react";import streamsaver from "streamsaver";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../emui/enclaves/types";
import { CopyButton } from "./CopyButton";
import {ServiceInfo} from "enclave-manager-sdk/build/api_container_service_pb";
import {isDefined, stripAnsi} from "../utils";
import {LogLineMessage} from "./enclaves/logs/types";
import { Timestamp } from "@bufbuild/protobuf";
import {DateTime} from "luxon";

type CopyLogsButtonProps = {
    logsFileName:string,
    enclave:EnclaveFullInfo,
    service?:ServiceInfo,
    logsToDownload: LogLineMessage[];
};

export const CopyLogsButton = ({ logsFileName, enclave, service, logsToDownload }: CopyLogsButtonProps) => {
    const kurtosisClient = useKurtosisClient();
    const [isLoading, setIsLoading] = useState(false);
    const [logLinesToDownload, setLogLinesToDownload] = useState(logsToDownload);

    const serviceLogLineToLogLineMessage = (lines: string[], timestamp?: Timestamp): LogLineMessage[] => {
        return lines.map((line) => ({
            message: line,
            timestamp: isDefined(timestamp) ? DateTime.fromJSDate(timestamp?.toDate()) : undefined,
        }));
    };

    const handleCopyLogsClick = async () => {
        setIsLoading(true);
        const writableStream = streamsaver.createWriteStream(logsFileName || "logs.txt");
        const writer = writableStream.getWriter();

        if (service){
            console.log("pulling logs")
            const abortController = new AbortController();
            for await (const lineGroup of await kurtosisClient.getServiceLogs(abortController, enclave, [service], false, 0, true)) {
                const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
                if (!isDefined(lineGroupForService)) continue;
                const parsedLogLines = serviceLogLineToLogLineMessage(lineGroupForService.line, lineGroupForService.timestamp);
                console.log("writing logs")
                setLogLinesToDownload((logLinesToDownload) => [...logLinesToDownload, ...parsedLogLines]);
            }

            try {
                console.log("downloading logs")
                await writer.write(logLinesToDownload.map(({message}) => message)
                    .filter(isDefined)
                    .map(stripAnsi)
                    .join("\n"));
            } catch(err) {
                console.error(err)
            }
            await writer.close();
        }

        if(logsToDownload) {
            try {
                console.log("copying logs")
                await writer.write(getLogsValue())
            } catch(err) {
                console.error(err)
            }
            await writer.close();
        }
        await writer.close();
        console.log("finished downloading logs")
        setLogLinesToDownload(() => [])
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
        <CopyButton
            contentName={"logs"}
            size={"sm"}
            // isDisabled={logLinesToDownload.length === 0}
            isIconButton
            aria-label={"Copy logs"}
            color={"gray.100"}
            isLoading={isLoading}
            onClick={handleCopyLogsClick}
        />
    );
};

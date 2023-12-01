import { useState } from "react";import streamsaver from "streamsaver";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { DownloadButton } from "../../CopyButton";
import {ServiceInfo} from "enclave-manager-sdk/build/api_container_service_pb";
import {isDefined, stripAnsi} from "../utils";

type DownloadLogsButtonProps = {
    enclave:EnclaveFullInfo,
    service?:ServiceInfo,
};

export const DownloadLogsButton = ({ enclave, service }: DownloadLogsButtonProps) => {
    const kurtosisClient = useKurtosisClient();
    const [isLoading, setIsLoading] = useState(false);
    const [logLinesToDownload, setLogLinesToDownload] = useState(propsLogLines);

    const handleDownloadClick = async () => {
        setIsLoading(true);
        const abortController = new AbortController();
        const writableStream = streamsaver.createWriteStream(logsFileName || "logs.txt");
        const writer = writableStream.getWriter();

        if (service) {
            console.log("pulling logs")
            for await (const lineGroup of await kurtosisClient.getServiceLogs(abortController, enclave, [service], false, 0, true)) {
                const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
                if (!isDefined(lineGroupForService)) continue;
                const parsedLogLines = serviceLogLineToLogLineMessage(lineGroupForService.line, lineGroupForService.timestamp);
                console.log("writing logs")
                setLogLinesToDownload((logLinesToDownload) => [...logLinesToDownload, ...parsedLogLines]);
            }
        } else {
            setLogLinesToDownload(() => [...logLines])
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
        console.log("finished downloading logs")
        setLogLinesToDownload(() => [])
        setIsLoading(false);
    };

    return (
        <CopyButton
            contentName={"logs"}
            valueToCopy={getLogsValue}
            size={"sm"}
            isDisabled={logLines.length === 0}
            isIconButton
            aria-label={"Copy logs"}
            color={"gray.100"}
        />
    );
};

import { Timestamp } from "@bufbuild/protobuf";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { DateTime } from "luxon";
import { useEffect, useState } from "react";
import { useKurtosisClient } from "../../../../../client/enclaveManager/KurtosisClientContext";
import { LogLineProps } from "../../../../../components/enclaves/logs/LogLine";
import { LogViewer } from "../../../../../components/enclaves/logs/LogViewer";
import { isDefined } from "../../../../../utils";
import { EnclaveFullInfo } from "../../../types";

const serviceLogLineToLogLineProps = (lines: string[], timestamp?: Timestamp): LogLineProps[] => {
  return lines.map((line) => ({
    message: line,
    timestamp: isDefined(timestamp) ? DateTime.fromJSDate(timestamp?.toDate()) : undefined,
  }));
};

type ServiceLogsProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

// @ts-ignore
export async function reTryCatch(callback, times: number = 1, isRetry: boolean = false) {
  try {
    return await callback(isRetry);
  } catch (error) {
    if (times > 0) {
      console.log("retrying another times ", times);
      return await reTryCatch(callback, times - 1, true);
    } else {
      console.log("retry: giving up and throwing error");
      throw error;
    }
  }
}

export const ServiceLogs = ({ enclave, service }: ServiceLogsProps) => {
  const kurtosisClient = useKurtosisClient();
  const [logLines, setLogLines] = useState<LogLineProps[]>([]);

  useEffect(() => {
    let canceled = false;
    const abortController = new AbortController();
    setLogLines([]);
    const callback = async (isRetry: boolean) => {
      // TODO: when we have a way to track where we left off, we don't have to clear and re-read everything
      if(isRetry) setLogLines([]);
      try {
        console.log("Creating a new logging stream");
        for await (const lineGroup of await kurtosisClient.getServiceLogs(
          abortController,
          enclave,
          [service],
        )) {
          if (canceled) return;
          const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
          if (!isDefined(lineGroupForService)) continue;
          const parsedLines = serviceLogLineToLogLineProps(lineGroupForService.line, lineGroupForService.timestamp);
          setLogLines((logLines) => [...logLines, ...parsedLines]);
        }
      } catch (error: any) {
        if (canceled) {
          console.log("The logging stream was successfully canceled (not a real error)", error);
          return;
        }
        console.error("An unhandled error occurred while streaming logs", error);
        throw error;
      }
    };
    reTryCatch(callback, 10000);
    return () => {
      canceled = true;
      abortController.abort();
    };
  }, [enclave, service, kurtosisClient]);

  const logsFileName = `${enclave.name}--${service.name}-logs.txt`;
  return <LogViewer logLines={logLines} logsFileName={logsFileName} />;
};

import { Timestamp } from "@bufbuild/protobuf";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { assertDefined, isDefined, LogLineMessage, LogViewer, stringifyError } from "kurtosis-ui-components";
import { DateTime } from "luxon";
import { useCallback, useEffect, useState } from "react";
import { useKurtosisClient } from "../../../../../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../../../types";
import React from "react";

const serviceLogLineToLogLineMessage = (lines: string[], timestamp?: Timestamp): LogLineMessage[] => {
  return lines.map((line) => ({
    message: line,
    timestamp: isDefined(timestamp) ? DateTime.fromJSDate(timestamp?.toDate()) : undefined,
  }));
};

type ServiceLogsProps = {
  enclave: EnclaveFullInfo;
  service: ServiceInfo;
};

export async function reTryCatch<R>(
  callback: (isRetry: boolean) => Promise<R>,
  times: number = 1,
  isRetry: boolean = false,
): Promise<R> {
  try {
    return await callback(isRetry);
  } catch (error) {
    if (times > 0) {
      console.info(`retrying another ${times} times`);
      return await reTryCatch(callback, times - 1, true);
    } else {
      console.info("retry: giving up and throwing error");
      throw error;
    }
  }
}

export const ServiceLogs = ({ enclave, service }: ServiceLogsProps) => {
  const kurtosisClient = useKurtosisClient();
  const [logLines, setLogLines] = useState<LogLineMessage[]>([]);

  const handleGetAllLogs = useCallback(
    async function* () {
      const abortController = new AbortController();
      const logs = await kurtosisClient.getServiceLogs(abortController, enclave, [service], false, 0, true);
      try {
        for await (const lineGroup of logs) {
          const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
          assertDefined(
            lineGroupForService,
            `Log line response included a line group withouth service ${
              service.serviceUuid
            }: ${lineGroup.toJsonString()}`,
          );
          const parsedLogLines = serviceLogLineToLogLineMessage(
            lineGroupForService.line,
            lineGroupForService.timestamp,
          );
          yield parsedLogLines.map((line) => line.message || "").join("\n");
        }
      } catch (err: any) {
        console.error(stringifyError(err));
      }
    },
    [kurtosisClient, enclave, service],
  );

  useEffect(() => {
    let canceled = false;
    const abortController = new AbortController();
    setLogLines([]);
    const callback = async (isRetry: boolean) => {
      // TODO: when we have a way to track where we left off, we don't have to clear and re-read everything
      if (isRetry) setLogLines([]);
      console.info("Created a new logging stream");
      try {
        for await (const lineGroup of await kurtosisClient.getServiceLogs(abortController, enclave, [service])) {
          if (canceled) return;
          const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
          if (!isDefined(lineGroupForService)) continue;
          const parsedLines = serviceLogLineToLogLineMessage(lineGroupForService.line, lineGroupForService.timestamp);
          setLogLines((logLines) => [...logLines, ...parsedLines]);
        }
      } catch (error: any) {
        if (canceled) {
          console.info("The logging stream was successfully canceled (not an error)", error);
          return;
        }
        console.error("An unhandled error occurred while streaming logs", error);
        throw error;
      }
    };
    reTryCatch(callback, 25);
    return () => {
      canceled = true;
      abortController.abort();
    };
  }, [enclave, service, kurtosisClient]);

  const logsFileName = `${enclave.name}--${service.name}-logs.txt`;
  return <LogViewer logLines={logLines} logsFileName={logsFileName} searchEnabled onGetAllLogs={handleGetAllLogs} />;
};

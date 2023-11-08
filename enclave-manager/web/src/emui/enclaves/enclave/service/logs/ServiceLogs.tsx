import { Timestamp } from "@bufbuild/protobuf";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { DateTime } from "luxon";
import { useEffect, useState } from "react";
import { useKurtosisClient } from "../../../../../client/enclaveManager/KurtosisClientContext";
import { LogLineProps } from "../../../../../components/enclaves/logs/LogLine";
import { LogViewer } from "../../../../../components/enclaves/logs/LogViewer";
import { isDefined, stringifyError } from "../../../../../utils";
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

export const ServiceLogs = ({ enclave, service }: ServiceLogsProps) => {
  const kurtosisClient = useKurtosisClient();
  const [logLines, setLogLines] = useState<LogLineProps[]>([]);

  useEffect(() => {
    let cancelled = false;
    const abortController = new AbortController();
    (async () => {
      setLogLines([]);
      try {
        for await (const lineGroup of await kurtosisClient.getServiceLogs(abortController, enclave, [service])) {
          if (cancelled) {
            return;
          }
          const lineGroupForService = lineGroup.serviceLogsByServiceUuid[service.serviceUuid];
          if (!isDefined(lineGroupForService)) {
            continue;
          }
          const parsedLines = serviceLogLineToLogLineProps(lineGroupForService.line, lineGroupForService.timestamp);
          setLogLines((logLines) => [...logLines, ...parsedLines]);
        }
      } catch (error: any) {
        if (cancelled) {
          return;
        }
        setLogLines((logLines) => [...logLines, { message: `Error: ${stringifyError(error)}`, status: "error" }]);
      }
    })();
    return () => {
      cancelled = true;
      abortController.abort();
    };
  }, [enclave, service, kurtosisClient]);

  return <LogViewer logLines={logLines} />;
};

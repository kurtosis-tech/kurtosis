import { Flex, Progress } from "@chakra-ui/react";
import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { useEffect, useState } from "react";
import { Location, useLocation, useNavigate } from "react-router-dom";
import { Virtuoso } from "react-virtuoso";
import { LogLine, LogLineProps } from "../../../../components/enclaves/logs/LogLine";
import { isAsyncIterable } from "../../../../utils";
import { EnclaveFullInfo } from "../../types";
import { RunStarlarkResolvedType } from "../action";

const LOG_START_INTERPRETTING = "Interpreting Starlark code - execution will begin shortly";
const LOG_START_VALIDATION = "Starting validation";
const LOG_DOWNLOADING = "Validating Starlark code and downloading container images - execution will begin shortly";
const LOG_STARTING_EXECUTION = "Starting execution";

export function starlarkResponseLineToLogLineProps(l: StarlarkRunResponseLine): LogLineProps {
  switch (l.runResponseLine.case) {
    case "instruction":
      return { message: l.runResponseLine.value.executableInstruction };
    case "progressInfo":
      return { message: l.runResponseLine.value.currentStepInfo[l.runResponseLine.value.currentStepNumber] };
    case "instructionResult":
      return { message: l.runResponseLine.value.serializedInstructionResult };
    case "error":
      return { message: l.runResponseLine.value.error.value?.errorMessage || "Unknown error", status: "error" };
    case "runFinishedEvent":
      return { message: l.runResponseLine.value.isRunSuccessful ? "Script completed" : "Script failed" };
    default:
      return { message: `Unknown line: ${l.toJsonString()}` };
  }
}

type EnclaveLogsProps = {
  enclave: EnclaveFullInfo;
};

export const EnclaveLogs = ({ enclave }: EnclaveLogsProps) => {
  const navigator = useNavigate();
  const location = useLocation() as Location<RunStarlarkResolvedType | undefined>;
  const [progress, setProgress] = useState(0);
  const [finalStatus, setFinalStatus] = useState<"ok" | "failed">();
  const [logLines, setLogLines] = useState<StarlarkRunResponseLine[]>([]);

  useEffect(() => {
    (async () => {
      if (location.state && isAsyncIterable(location.state.logs)) {
        setLogLines([]);
        setFinalStatus(undefined);
        setProgress(0);
        for await (const line of location.state.logs) {
          setLogLines((logLines) => [...logLines, line]);
          // Progress will be 10% validation, 90% execution
          if (line.runResponseLine.case === "progressInfo") {
            const totalSteps = line.runResponseLine.value.totalSteps;
            const completedSteps = Math.max(line.runResponseLine.value.currentStepNumber - 1, 0);
            const progress = (completedSteps / totalSteps) * 100;
            setProgress(progress);
          }

          if (line.runResponseLine.case === "runFinishedEvent") {
            setFinalStatus(line.runResponseLine.value.isRunSuccessful ? "ok" : "failed");
            setProgress(100);
          }

          console.log(line.runResponseLine.value);
        }
      } else {
        navigator(`/enclave/${enclave.shortenedUuid}/overview`);
      }
    })();
  }, [location]);

  return (
    <Flex display={"column"} bg={"gray.800"}>
      <Virtuoso
        style={{ height: "400px" }}
        data={logLines}
        itemContent={(index, line) => <LogLine {...starlarkResponseLineToLogLineProps(line)} />}
      />
      <Progress value={progress} isIndeterminate={progress === 0} height={"4px"} colorScheme={"kurtosisGreen"} />
    </Flex>
  );
};

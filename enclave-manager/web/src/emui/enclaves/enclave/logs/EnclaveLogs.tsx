import { CircularProgress, Icon } from "@chakra-ui/react";
import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { useEffect, useState } from "react";
import { FiCheck, FiX } from "react-icons/fi";
import { Location, useLocation, useNavigate, useRevalidator } from "react-router-dom";
import { LogLineProps } from "../../../../components/enclaves/logs/LogLine";
import { LogViewer } from "../../../../components/enclaves/logs/LogViewer";
import { isAsyncIterable, stringifyError } from "../../../../utils";
import { EnclaveFullInfo } from "../../types";
import { RunStarlarkResolvedType } from "../action";

// These are the stages we want to catch and handle in the UI
type EnclaveLogStage =
  | { stage: "waiting" }
  | { stage: "validating" }
  | { stage: "executing"; step: number; totalSteps: number }
  | { stage: "done"; totalSteps: number | null }
  | { stage: "failed" };

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
    case "info":
      return { message: l.runResponseLine.value.infoMessage };
    default:
      return { message: `Unknown line: ${l.toJsonString()}` };
  }
}

type EnclaveLogsProps = {
  enclave: EnclaveFullInfo;
};

export const EnclaveLogs = ({ enclave }: EnclaveLogsProps) => {
  const navigator = useNavigate();
  const revalidator = useRevalidator();
  const location = useLocation() as Location<RunStarlarkResolvedType | undefined>;
  const [progress, setProgress] = useState<EnclaveLogStage>({ stage: "waiting" });
  const [logLines, setLogLines] = useState<LogLineProps[]>([]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (location.state && isAsyncIterable(location.state.logs)) {
        setLogLines([]);
        setProgress({ stage: "waiting" });
        try {
          for await (const line of location.state.logs) {
            if (cancelled) {
              return;
            }
            const parsedLine = starlarkResponseLineToLogLineProps(line);
            setLogLines((logLines) => [...logLines, parsedLine]);
            setProgress((oldProgress) => {
              if (line.runResponseLine.case === "progressInfo") {
                if (oldProgress.stage === "waiting") {
                  return {
                    stage: "validating",
                  };
                }
                if (parsedLine.message === LOG_STARTING_EXECUTION || oldProgress.stage === "executing") {
                  return {
                    stage: "executing",
                    totalSteps: line.runResponseLine.value.totalSteps,
                    step: line.runResponseLine.value.currentStepNumber,
                  };
                }
              }
              if (line.runResponseLine.case === "runFinishedEvent") {
                revalidator.revalidate();
                return line.runResponseLine.value.isRunSuccessful
                  ? { stage: "done", totalSteps: oldProgress.stage === "executing" ? oldProgress.totalSteps : null }
                  : { stage: "failed" };
              }
              return oldProgress;
            });

            console.log(line.runResponseLine.value);
          }
        } catch (error: any) {
          if (cancelled) {
            return;
          }
          setLogLines((logLines) => [...logLines, { message: `Error: ${stringifyError(error)}`, status: "error" }]);
          revalidator.revalidate();
        }
      } else {
        navigator(`/enclave/${enclave.shortenedUuid}/overview`);
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location, enclave.shortenedUuid, navigator, revalidator.revalidate]);

  const progressPercent =
    progress.stage === "validating"
      ? "indeterminate"
      : progress.stage === "failed"
      ? "failed"
      : progress.stage === "executing"
      ? (100 * progress.step + 1) / (progress.totalSteps + 1)
      : progress.stage === "done"
      ? 100
      : 0;

  return (
    <LogViewer
      logLines={logLines}
      progressPercent={progressPercent}
      ProgressWidget={<ProgressSummary progress={progress} />}
      logsFileName={`${enclave.name.replaceAll(/\s+/g, "_")}-logs.txt`}
    />
  );
};

type ProgressSummaryProps = {
  progress: EnclaveLogStage;
};

const ProgressSummary = ({ progress }: ProgressSummaryProps) => {
  return (
    <>
      {progress.stage === "waiting" && "Waiting"}
      {progress.stage === "validating" && "Validating"}
      {progress.stage === "executing" && (
        <>
          <CircularProgress
            size={"18px"}
            value={(100 * progress.step + 1) / (progress.totalSteps + 1)}
            color={"kurtosisGreen.400"}
          />
          <span>
            {progress.step} / {progress.totalSteps}
          </span>
        </>
      )}
      {progress.stage === "done" && (
        <>
          <Icon as={FiCheck} size={"18px"} color={"kurtosisGreen.400"} />
          <span>
            {progress.totalSteps} / {progress.totalSteps}
          </span>
        </>
      )}
      {progress.stage === "failed" && (
        <>
          <Icon as={FiX} size={"18px"} color={"red.400"} />
          <span>Failed</span>
        </>
      )}
    </>
  );
};

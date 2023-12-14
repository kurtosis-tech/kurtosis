import { ButtonGroup, CircularProgress, Flex, Icon, Tag } from "@chakra-ui/react";
import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { AppPageLayout, isAsyncIterable, LogLineMessage, LogViewer, stringifyError } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { FiCheck, FiX } from "react-icons/fi";
import { Location, useLocation, useNavigate } from "react-router-dom";
import { EditEnclaveButton } from "../../components/EditEnclaveButton";
import { DeleteEnclavesButton } from "../../components/widgets/DeleteEnclavesButton";
import { useEnclavesContext } from "../../EnclavesContext";
import { useEnclaveFromParams } from "../EnclaveRouteContext";

// These are the stages we want to catch and handle in the UI
type EnclaveLogStage =
  | { stage: "waiting" }
  | { stage: "validating" }
  | { stage: "executing"; step: number; totalSteps: number }
  | { stage: "done"; totalSteps: number | null }
  | { stage: "failed" };

const LOG_STARTING_EXECUTION = "Starting execution";

export function starlarkResponseLineToLogLineMessage(l: StarlarkRunResponseLine): LogLineMessage {
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

export const EnclaveLogs = () => {
  const enclave = useEnclaveFromParams();
  const { refreshServices, refreshFilesAndArtifacts, refreshStarlarkRun, updateStarlarkFinishedInEnclave } =
    useEnclavesContext();
  const navigator = useNavigate();
  const location = useLocation() as Location<{ logs: AsyncIterable<StarlarkRunResponseLine> }>;
  const [progress, setProgress] = useState<EnclaveLogStage>({ stage: "waiting" });
  const [logLines, setLogLines] = useState<LogLineMessage[]>([]);

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
            const parsedLine = starlarkResponseLineToLogLineMessage(line);
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
                return line.runResponseLine.value.isRunSuccessful
                  ? { stage: "done", totalSteps: oldProgress.stage === "executing" ? oldProgress.totalSteps : null }
                  : { stage: "failed" };
              }
              return oldProgress;
            });
            if (line.runResponseLine.case === "runFinishedEvent") {
              await Promise.all([
                refreshStarlarkRun(enclave),
                refreshServices(enclave),
                refreshFilesAndArtifacts(enclave),
              ]);
            }
          }
        } catch (error: any) {
          if (cancelled) {
            return;
          }
          setLogLines((logLines) => [...logLines, { message: `Error: ${stringifyError(error)}`, status: "error" }]);
          await Promise.all([refreshStarlarkRun(enclave), refreshServices(enclave), refreshFilesAndArtifacts(enclave)]);
        } finally {
          updateStarlarkFinishedInEnclave(enclave);
        }
      } else {
        navigator(`/enclave/${enclave.shortenedUuid}/overview`);
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location, enclave.shortenedUuid, navigator]);

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
    <AppPageLayout preventPageScroll>
      <LogViewer
        logLines={logLines}
        progressPercent={progressPercent}
        copyLogsEnabled
        ProgressWidget={
          <Flex justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
            <ProgressSummary progress={progress} />
            <ButtonGroup>
              <DeleteEnclavesButton enclaves={[enclave]} variant={"ghost"} size={"md"} />
              <EditEnclaveButton enclave={enclave} variant={"ghost"} size={"md"} />
            </ButtonGroup>
          </Flex>
        }
        logsFileName={`${enclave.name.replaceAll(/\s+/g, "_")}-logs.txt`}
      />
    </AppPageLayout>
  );
};

type ProgressSummaryProps = {
  progress: EnclaveLogStage;
};

const ProgressSummary = ({ progress }: ProgressSummaryProps) => {
  return (
    <Tag
      variant={"progress"}
      p={"0 16px"}
      h={"40px"}
      fontSize={"md"}
      colorScheme={progress.stage === "done" ? "green" : progress.stage === "failed" ? "red" : "blue"}
    >
      <Flex gap={"8px"} alignItems={"center"}>
        {progress.stage === "waiting" && "Waiting"}
        {progress.stage === "validating" && "Validating"}
        {progress.stage === "executing" && (
          <>
            <CircularProgress size={"18px"} value={(100 * progress.step + 1) / (progress.totalSteps + 1)} />
            <span>
              {progress.step} / {progress.totalSteps}
            </span>
          </>
        )}
        {progress.stage === "done" && (
          <>
            <Icon as={FiCheck} size={"18px"} />
            <span>
              {progress.totalSteps} / {progress.totalSteps}
            </span>
          </>
        )}
        {progress.stage === "failed" && (
          <>
            <Icon as={FiX} size={"18px"} />
            <span>Failed</span>
          </>
        )}
      </Flex>
    </Tag>
  );
};

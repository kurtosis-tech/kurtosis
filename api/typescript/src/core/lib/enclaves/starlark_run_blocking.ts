import {Readable} from "stream";
import {
    StarlarkExecutionError,
    StarlarkInstruction,
    StarlarkInterpretationError, StarlarkRunResponseLine, StarlarkValidationError
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

const STARLARK_RUN_OUTPUT_LINE_SPLIT = "\n"

export class StarlarkRunResult {
    constructor(
        public readonly runOutput: string,
        public readonly instructions: Array<StarlarkInstruction>,
        public readonly interpretationError: StarlarkInterpretationError | undefined,
        public readonly validationErrors: Array<StarlarkValidationError>,
        public readonly executionError: StarlarkExecutionError | undefined
    ){}
}

export async function readStreamContentUntilClosed(responseLines: Readable): Promise<StarlarkRunResult> {
    let scriptOutput = ""
    let interpretationError: StarlarkInterpretationError | undefined
    let validationErrors: Array<StarlarkValidationError> = []
    let executionError: StarlarkExecutionError | undefined
    let instructions: Array<StarlarkInstruction> = []

    return new Promise((resolve, error) => {
        responseLines.on('data', (responseLine: StarlarkRunResponseLine) => {
            if (responseLine.getInstruction() !== undefined) {
                instructions.push(responseLine.getInstruction()!)
            } else if (responseLine.getInstructionResult() !== undefined) {
                scriptOutput += responseLine.getInstructionResult()?.getSerializedInstructionResult() + STARLARK_RUN_OUTPUT_LINE_SPLIT
            } else if (responseLine.getError() !== undefined) {
                if (responseLine.getError()?.getInterpretationError() !== undefined) {
                    interpretationError = responseLine.getError()?.getInterpretationError()
                } else if (responseLine.getError()?.getValidationError() !== undefined) {
                    validationErrors.push(responseLine.getError()!.getValidationError()!)
                } else if (responseLine.getError()?.getExecutionError() !== undefined) {
                    executionError = responseLine.getError()?.getExecutionError()
                }
            } else if (responseLine.getRunFinishedEvent() !== undefined) {
                let runFinishedEvent = responseLine.getRunFinishedEvent()!
                if (runFinishedEvent.getIsrunsuccessful() && runFinishedEvent.getSerializedOutput() != "")  {
                        scriptOutput += runFinishedEvent.getSerializedOutput() + STARLARK_RUN_OUTPUT_LINE_SPLIT
                }
            }
        })
        responseLines.on('error', function () {
            if (!responseLines.destroyed) {
                responseLines.destroy();
                error(new Error("Unexpected error"))
            }
        });
        responseLines.on('end', function () {
            if (!responseLines.destroyed) {
                responseLines.destroy();
                resolve(new StarlarkRunResult(scriptOutput, instructions, interpretationError, validationErrors, executionError))
            }
        });
    })
}

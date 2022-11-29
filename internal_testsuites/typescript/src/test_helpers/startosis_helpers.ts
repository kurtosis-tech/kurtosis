import {
    StarlarkExecutionError,
    StarlarkRunResponseLine,
    StarlarkInstruction,
    StarlarkInterpretationError,
    StarlarkValidationError
} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import {Readable} from "stream";

const NEWLINE_CHAR = "\n"

export function readStreamContentUntilClosed(responseLines: Readable): Promise<[
    string,
    Array<StarlarkInstruction>,
    StarlarkInterpretationError | undefined,
    Array<StarlarkValidationError>,
    StarlarkExecutionError | undefined
]> {
    let scriptOutput = ""
    let interpretationError: StarlarkInterpretationError | undefined
    let validationErrors: Array<StarlarkValidationError> = []
    let executionError: StarlarkExecutionError | undefined
    let instructions: Array<StarlarkInstruction> = []

    return new Promise(resolve => {
        responseLines.on('data', (responseLine: StarlarkRunResponseLine) => {
            if (responseLine.getInstruction() !== undefined) {
                instructions.push(responseLine.getInstruction()!)
            } else if (responseLine.getInstructionResult() !== undefined) {
                scriptOutput += responseLine.getInstructionResult()?.getSerializedInstructionResult() + NEWLINE_CHAR
            } else if (responseLine.getError() !== undefined) {
                if (responseLine.getError()?.getInterpretationError() !== undefined) {
                    interpretationError = responseLine.getError()?.getInterpretationError()
                } else if (responseLine.getError()?.getValidationError() !== undefined) {
                    validationErrors.push(responseLine.getError()!.getInterpretationError()!)
                } else if (responseLine.getError()?.getExecutionError() !== undefined) {
                    executionError = responseLine.getError()?.getExecutionError()
                }
            }
        })
        responseLines.on('error', function () {
            if (!responseLines.destroyed) {
                responseLines.destroy();
                throw new Error("Unexpected error");
            }
        });
        responseLines.on('end', function () {
            if (!responseLines.destroyed) {
                responseLines.destroy();
                resolve([scriptOutput, instructions, interpretationError, validationErrors, executionError])
            }
        });
    })
}

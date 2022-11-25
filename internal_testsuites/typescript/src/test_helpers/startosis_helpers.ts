import {
    KurtosisExecutionError, KurtosisExecutionResponseLine,
    KurtosisInstruction,
    KurtosisInterpretationError, KurtosisValidationError, KurtosisValidationErrors
} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import {Readable} from "stream";


export function generateScriptOutput(instructions: Array<KurtosisInstruction>): string {
    let scriptOutput = "";
    instructions.forEach((instruction) => {
        if (instruction.hasInstructionResult()) {
            scriptOutput += instruction.getInstructionResult() + "\n"
        }
    })
    return scriptOutput
}

export function readStreamContentUntilClosed(responseLines: Readable): Promise<[
    KurtosisInterpretationError | undefined,
    Array<KurtosisValidationError>,
    KurtosisExecutionError | undefined,
    Array<KurtosisInstruction>
]> {
    let interpretationError: KurtosisInterpretationError | undefined
    let validationErrors: Array<KurtosisValidationError> = []
    let executionError: KurtosisExecutionError | undefined
    let instructions: Array<KurtosisInstruction> = []

    return new Promise(resolve => {
        responseLines.on('data', (responseLine: KurtosisExecutionResponseLine) => {
            if (responseLine.getInstruction() !== undefined) {
                instructions.push(responseLine.getInstruction()!)
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
                resolve([interpretationError, validationErrors, executionError, instructions])
            }
        });
    })
}

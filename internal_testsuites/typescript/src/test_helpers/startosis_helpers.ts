import {KurtosisInstruction} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";


export function generateScriptOutput(instructions: Array<KurtosisInstruction>): string {
    let scriptOutput = "";
    instructions.forEach((instruction) => {
        if (instruction.hasInstructionResult()) {
            scriptOutput += instruction.getInstructionResult()
        }
    })
    return scriptOutput
}

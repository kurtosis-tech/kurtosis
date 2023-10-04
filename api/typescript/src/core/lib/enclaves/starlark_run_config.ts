import {KurtosisFeatureFlag} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

type StarlarkRunConfigOption = (starlarkRunConfig: StarlarkRunConfig) => void;

const DEFAULT_RELATIVE_PATH_TO_MAIN_FILE = ""
const DEFAULT_MAIN_FUNCTION_NAME = ""
const DEFAULT_SERIALIZED_PARAMS = "{}"
const DEFAULT_DRY_RUN = false
const DEFAULT_PARALLELISM = 4
const DEFAULT_EXPERIMENTAL_FEATURE_FLAGS = Array<KurtosisFeatureFlag>()
const DEFAULT_CLOUD_INSTANCE_ID = ""
const DEFAULT_CLOUD_USER_ID = ""

export class StarlarkRunConfig {
    public relativePathToMainFile: string
    public mainFunctionName: string
    public serializedParams: string
    public dryRun: boolean
    public parallelism: number
    public experimentalFeatureFlags: Array<KurtosisFeatureFlag>
    public cloudInstanceId: string
    public cloudUserId: string

    constructor(...options: StarlarkRunConfigOption[]) {
        this.relativePathToMainFile = DEFAULT_RELATIVE_PATH_TO_MAIN_FILE
        this.mainFunctionName = DEFAULT_MAIN_FUNCTION_NAME
        this.serializedParams = DEFAULT_SERIALIZED_PARAMS
        this.dryRun = DEFAULT_DRY_RUN
        this.parallelism = DEFAULT_PARALLELISM
        this.experimentalFeatureFlags = DEFAULT_EXPERIMENTAL_FEATURE_FLAGS
        this.cloudInstanceId = DEFAULT_CLOUD_INSTANCE_ID
        this.cloudUserId = DEFAULT_CLOUD_USER_ID

        for (const option of options) {
            option(this)
        }
    }

    public static WithRelativePathToMainFile(relativePathToMainFile: string): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.relativePathToMainFile = relativePathToMainFile
        }
    }

    public static WithMainFunctionName(mainFunctionName: string): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.mainFunctionName = mainFunctionName
        }
    }

    public static WithSerializedParams(serializedParams: string): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.serializedParams = serializedParams
        }
    }

    public static WithDryRun(dryRun: boolean): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.dryRun = dryRun
        }
    }

    public static WithParallelism(parallelism: number): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.parallelism = parallelism
        }
    }

    public static WithExperimentalFeatureFlags(experimentalFeatureFlags: Array<KurtosisFeatureFlag>): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.experimentalFeatureFlags = experimentalFeatureFlags
        }
    }

    public static WithCloudInstanceId(cloudInstanceId: string): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.cloudInstanceId = cloudInstanceId
        }
    }

    public static WithCloudUserID(cloudUserId: string): StarlarkRunConfigOption {
        return (config: StarlarkRunConfig): void => {
            config.cloudUserId = cloudUserId
        }
    }
}
import {
  GetServicesResponse,
  GetStarlarkRunResponse,
  ListFilesArtifactNamesAndUuidsResponse,
} from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { RemoveFunctions } from "kurtosis-ui-components";
import { Result } from "true-myth";

export type EnclaveFullInfo = RemoveFunctions<EnclaveInfo> & {
  starlarkRun?: Result<RemoveFunctions<GetStarlarkRunResponse>, string>;
  services?: Result<RemoveFunctions<GetServicesResponse>, string>;
  filesAndArtifacts?: Result<RemoveFunctions<ListFilesArtifactNamesAndUuidsResponse>, string>;
  initialSubmissionData?: Record<string, any>;
  initialPackageId?: string;
};

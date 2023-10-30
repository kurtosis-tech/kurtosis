import {
  GetServicesResponse,
  GetStarlarkRunResponse,
  ListFilesArtifactNamesAndUuidsResponse,
} from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { RemoveFunctions } from "../../utils/types";

export type EnclaveFullInfo = RemoveFunctions<EnclaveInfo> & {
  starlarkRun: RemoveFunctions<GetStarlarkRunResponse>;
  services: RemoveFunctions<GetServicesResponse>;
  filesAndArtifacts: RemoveFunctions<ListFilesArtifactNamesAndUuidsResponse>;
};

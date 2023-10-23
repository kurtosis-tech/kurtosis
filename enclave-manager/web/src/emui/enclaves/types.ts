import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import {
  FilesArtifactNameAndUuid,
  GetServicesResponse,
  GetStarlarkRunResponse,
  ListFilesArtifactNamesAndUuidsResponse,
} from "enclave-manager-sdk/build/api_container_service_pb";

type NonFunctionKeyNames<T> = Exclude<
  {
    [key in keyof T]: T[key] extends Function ? never : key;
  }[keyof T],
  undefined
>;

type RemoveFunctions<T> = Pick<T, NonFunctionKeyNames<T>>;

export type EnclaveFullInfo = RemoveFunctions<EnclaveInfo> & {
  starlarkRun: RemoveFunctions<GetStarlarkRunResponse>;
  services: RemoveFunctions<GetServicesResponse>;
  filesAndArtifacts: RemoveFunctions<ListFilesArtifactNamesAndUuidsResponse>;
};

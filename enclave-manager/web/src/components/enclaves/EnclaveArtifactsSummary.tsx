import { Button } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";

type EnclaveArtifactsSummaryProps = {
  artifacts: FilesArtifactNameAndUuid[];
};

export const EnclaveArtifactsSummary = ({ artifacts }: EnclaveArtifactsSummaryProps) => {
  return (
    <Button variant={"ghost"} size={"xs"}>
      {artifacts.length}
    </Button>
  );
};

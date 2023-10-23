import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { Button } from "@chakra-ui/react";

type EnclaveArtifactsSummaryProps = {
  artifacts: FilesArtifactNameAndUuid[];
};

export const EnclaveArtifactsSummary = ({ artifacts }: EnclaveArtifactsSummaryProps) => {
  return (
    <Button variant={"kurtosisGhost"} size={"xs"}>
      {artifacts.length}
    </Button>
  );
};

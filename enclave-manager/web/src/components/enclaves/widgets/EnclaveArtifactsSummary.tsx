import { Button, Tag } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "../../../utils";

type EnclaveArtifactsSummaryProps = {
  artifacts: FilesArtifactNameAndUuid[] | null;
};

export const EnclaveArtifactsSummary = ({ artifacts }: EnclaveArtifactsSummaryProps) => {
  if (!isDefined(artifacts)) {
    return <Tag>Unknown</Tag>;
  }

  return (
    <Button variant={"ghost"} size={"xs"}>
      {artifacts.length}
    </Button>
  );
};

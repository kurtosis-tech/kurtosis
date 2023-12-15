import { Tag, Text } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { isDefined } from "kurtosis-ui-components";

type EnclaveArtifactsSummaryProps = {
  artifacts: FilesArtifactNameAndUuid[] | null;
};

export const EnclaveArtifactsSummary = ({ artifacts }: EnclaveArtifactsSummaryProps) => {
  if (!isDefined(artifacts)) {
    return <Tag>Unknown</Tag>;
  }

  return (
    <Text fontWeight={"semibold"} fontSize={"xs"}>
      {artifacts.length}
    </Text>
  );
};

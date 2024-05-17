import { Button, ButtonGroup, Flex } from "@chakra-ui/react";
import { AppPageLayout, KurtosisAlert, PageTitle } from "kurtosis-ui-components";
import { useEffect, useMemo, useState } from "react";
import { useExperiments } from "../experiments/ExperimentsContext";
import { KurtosisUpgrader } from "./components/KurtosisUpgrader";
import { EnclavesTable } from "./components/tables/EnclavesTable";
import { CreateEnclaveButton } from "./components/widgets/CreateEnclaveButton";
import { DeleteEnclavesButton } from "./components/widgets/DeleteEnclavesButton";
import { useFullEnclaves } from "./EnclavesContext";
import { EnclaveFullInfo } from "./types";
import {BrowserRecommendator} from "./components/BrowserRecommendator";

export const EnclaveList = () => {
  const enclaves = useFullEnclaves();

  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  const { experiments } = useExperiments();
  const enclavesKey = useMemo(
    () =>
      enclaves.isErr
        ? "error"
        : enclaves.value
            .map((enclave) => enclave.shortenedUuid)
            .sort()
            .join("|"),
    [enclaves],
  );

  useEffect(() => {
    setSelectedEnclaves([]);
  }, [enclavesKey]);

  return (
    <AppPageLayout>
      <Flex pl={"6px"} pb={"16px"} alignItems={"center"} justifyContent={"space-between"}>
        <PageTitle>Enclaves</PageTitle>
        <Flex gap={"24px"} alignItems={"center"}>
          {selectedEnclaves.length > 0 && (
            <ButtonGroup isAttached variant={"kurtosisGroupOutline"} size={"sm"}>
              <Button variant={"kurtosisDisabled"} colorScheme={"gray"}>
                {selectedEnclaves.length} selected
              </Button>
              <DeleteEnclavesButton enclaves={selectedEnclaves} />
            </ButtonGroup>
          )}
          <CreateEnclaveButton />
        </Flex>
      </Flex>
      <Flex direction="column" pt={"24px"} width={"100%"} gap={8}>
        {experiments.enableCloudVersionUpgrade && <KurtosisUpgrader />}
        <BrowserRecommendator />
        {enclaves.isOk && (
          <EnclavesTable
            enclavesData={enclaves.value}
            selection={selectedEnclaves}
            onSelectionChange={setSelectedEnclaves}
          />
        )}
        {enclaves.isErr && <KurtosisAlert message={enclaves.error} />}
      </Flex>
    </AppPageLayout>
  );
};

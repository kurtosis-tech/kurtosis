import { Button, ButtonGroup, Flex } from "@chakra-ui/react";
import { useEffect, useMemo, useState } from "react";
import { AppPageLayout } from "../../components/AppLayout";
import { CreateEnclaveButton } from "../../components/enclaves/CreateEnclaveButton";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { useFullEnclaves } from "../EmuiAppContext";
import { EnclaveFullInfo } from "./types";

export const EnclaveList = () => {
  const enclaves = useFullEnclaves();

  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

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
      <Flex direction="column" gap={"24px"} width={"100%"}>
        <Flex justifyContent={"flex-end"}>
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

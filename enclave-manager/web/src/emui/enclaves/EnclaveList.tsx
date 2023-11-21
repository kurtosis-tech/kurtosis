import { Button, ButtonGroup, Flex } from "@chakra-ui/react";
import { useEffect, useMemo, useState } from "react";
import { CreateEnclaveButton } from "../../components/enclaves/CreateEnclaveButton";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { MAIN_APP_MAX_WIDTH, MAIN_APP_PADDING } from "../../components/theme/constants";
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
    <Flex maxWidth={MAIN_APP_MAX_WIDTH} p={MAIN_APP_PADDING} w={"100%"}>
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
    </Flex>
  );
};

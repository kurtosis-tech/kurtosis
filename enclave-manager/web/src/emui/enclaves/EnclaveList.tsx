import { Button, ButtonGroup, Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useState } from "react";
import { FiPlus, FiTrash2 } from "react-icons/fi";
import { useRouteLoaderData } from "react-router-dom";
import { Result, ResultNS } from "true-myth";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { isDefined } from "../../utils";
import { EnclaveFullInfo } from "./types";

export const enclavesLoader =
  (kurtosisClient: KurtosisClient) => async (): Promise<Result<EnclaveFullInfo[], string>> => {
    const enclavesResponse = await kurtosisClient.getEnclaves();
    if (enclavesResponse.isErr) {
      return Result.err(enclavesResponse.error.message || "Unknown api error");
    }
    const enclaves = Object.values(enclavesResponse.value.enclaveInfo);
    const [starlarkRuns, services, filesAndArtifacts] = await Promise.all([
      Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
      Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
      Promise.all(enclaves.map((enclave) => kurtosisClient.listFilesArtifactNamesAndUuids(enclave))),
    ]);

    const starlarkErrors = starlarkRuns.filter(ResultNS.isErr);
    const servicesErrors = services.filter(ResultNS.isErr);
    const filesAndArtifactErrors = filesAndArtifacts.filter(ResultNS.isErr);
    if (starlarkErrors.length + servicesErrors.length + filesAndArtifactErrors.length > 0) {
      return Result.err(
        `Starlark errors: ${
          starlarkErrors.length > 0 ? starlarkErrors.map((r) => r.error.message).join("\n") : "None"
        }\nServices errors: ${
          servicesErrors.length > 0 ? servicesErrors.map((r) => r.error.message).join("\n") : "None"
        }\nFiles and Artifacts errors: ${
          filesAndArtifactErrors.length > 0 ? filesAndArtifactErrors.map((r) => r.error.message).join("\n") : "None"
        }`,
      );
    }

    return Result.ok(
      enclaves.map((enclave, i) => ({
        ...enclave,
        // These values are never actually null because of the checking above
        starlarkRun: starlarkRuns[i].unwrapOr(null)!,
        services: services[i].unwrapOr(null)!,
        filesAndArtifacts: filesAndArtifacts[i].unwrapOr(null)!,
      })),
    );
  };

export const EnclaveList = () => {
  const maybeEnclaves = useRouteLoaderData("enclaves") as Awaited<ReturnType<ReturnType<typeof enclavesLoader>>>;

  const [enclaves, setEnclaves] = useState<EnclaveFullInfo[]>();
  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  return (
    <Flex direction="column">
      <Tabs variant={"soft-rounded"} colorScheme={"kurtosisGreen"}>
        <Flex justifyContent={"space-between"}>
          <TabList>
            <Tab>Enclaves</Tab>
          </TabList>
          <Flex gap={"24px"} alignItems={"center"}>
            {selectedEnclaves.length > 0 && (
              <ButtonGroup isAttached variant={"kurtosisGroupOutline"} size={"sm"}>
                <Button variant={"kurtosisDisabled"} colorScheme={"gray"}>
                  {selectedEnclaves.length} selected
                </Button>
                <Button colorScheme={"red"} leftIcon={<FiTrash2 />}>
                  Delete
                </Button>
              </ButtonGroup>
            )}
            <Button colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
              Create Enclave
            </Button>
          </Flex>
        </Flex>
        <TabPanels>
          <TabPanel>
            {isDefined(maybeEnclaves) && maybeEnclaves.isOk && (
              <EnclavesTable
                enclavesData={maybeEnclaves.value}
                selection={selectedEnclaves}
                onSelectionChange={setSelectedEnclaves}
              />
            )}
            {isDefined(maybeEnclaves) && maybeEnclaves.isErr && <KurtosisAlert message={maybeEnclaves.error} />}
            {!isDefined(maybeEnclaves) && (
              <Flex justifyContent={"center"} p={"20px"}>
                <Spinner size={"xl"} />
              </Flex>
            )}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};

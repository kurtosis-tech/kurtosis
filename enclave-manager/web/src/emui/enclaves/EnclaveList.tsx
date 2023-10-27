import { Button, ButtonGroup, Flex, Spinner, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { Suspense, useEffect, useState } from "react";
import { FiPlus } from "react-icons/fi";
import { ActionFunction, Await, defer, json, redirect, useRouteLoaderData } from "react-router-dom";
import { Result, ResultNS } from "true-myth";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { EnclavesTable } from "../../components/enclaves/tables/EnclavesTable";
import { DeleteEnclavesButton } from "../../components/enclaves/widgets/DeleteEnclavesButton";
import { KurtosisAlert } from "../../components/KurtosisAlert";
import { isDefined } from "../../utils";
import { EnclaveFullInfo } from "./types";

const loadEnclaves = async (kurtosisClient: KurtosisClient): Promise<Result<EnclaveFullInfo[], string>> => {
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

export type EnclavesLoaderResolved = {
  enclaves: Awaited<ReturnType<typeof loadEnclaves>>;
};

export const enclavesLoader = (kurtosisClient: KurtosisClient) => async () => {
  return defer({ enclaves: loadEnclaves(kurtosisClient) });
};

export const enclavesAction =
  (kurtosisClient: KurtosisClient): ActionFunction =>
  async ({ params, request }) => {
    const formData = await request.json();
    const intent = formData["intent"];
    if (intent === "delete") {
      const uuids = formData["enclaveUUIDs"];
      if (!isDefined(uuids)) {
        throw json({ message: "Missing enclaveUUIDs" }, { status: 400 });
      }
      console.log(uuids);
      await Promise.all(uuids.map((uuid: string) => kurtosisClient.destroy(uuid)));
      return redirect("/enclaves");
    } else {
      console.log("blep");
      throw json({ message: "Invalid intent" }, { status: 400 });
    }
  };

export const EnclaveList = () => {
  const { enclaves } = useRouteLoaderData("enclaves") as EnclavesLoaderResolved;

  return (
    <Suspense
      fallback={
        <Flex justifyContent={"center"} p={"20px"}>
          <Spinner size={"xl"} />
        </Flex>
      }
    >
      <Await resolve={enclaves} children={(enclaves) => <EnclaveListImpl enclaves={enclaves} />} />
    </Suspense>
  );
};

type EnclaveListImplProps = {
  enclaves: EnclavesLoaderResolved["enclaves"];
};

const EnclaveListImpl = ({ enclaves }: EnclaveListImplProps) => {
  const [selectedEnclaves, setSelectedEnclaves] = useState<EnclaveFullInfo[]>([]);

  useEffect(() => {
    setSelectedEnclaves([]);
  }, [enclaves]);

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
                <DeleteEnclavesButton enclaves={selectedEnclaves} />
              </ButtonGroup>
            )}
            <Button colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
              Create Enclave
            </Button>
          </Flex>
        </Flex>
        <TabPanels>
          <TabPanel>
            {enclaves.isOk && (
              <EnclavesTable
                enclavesData={enclaves.value}
                selection={selectedEnclaves}
                onSelectionChange={setSelectedEnclaves}
              />
            )}
            {enclaves.isErr && <KurtosisAlert message={enclaves.error} />}
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Flex>
  );
};

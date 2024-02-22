import { Flex, Grid, GridItem, Spinner } from "@chakra-ui/react";
import { FormatDateTime, isDefined, KurtosisAlert, TitledBox, ValueCard } from "kurtosis-ui-components";
import { DateTime } from "luxon";
import { FilesTable } from "../../components/tables/FilesTable";
import { ServicesTable } from "../../components/tables/ServicesTable";
import { EnclaveStatus } from "../../components/widgets/EnclaveStatus";
import { EnclaveFullInfo } from "../../types";

type EnclaveOverviewProps = {
  enclave: EnclaveFullInfo;
};

export const EnclaveOverview = ({ enclave }: EnclaveOverviewProps) => {
  const enclaveCreationDateTime = isDefined(enclave.creationTime)
    ? DateTime.fromJSDate(enclave.creationTime.toDate())
    : null;

  return (
    <Flex flexDirection={"column"} gap={"32px"}>
      <Grid templateColumns={"repeat(4, 1fr)"} gap={"32px"}>
        <GridItem>
          <ValueCard title={"Name"} value={enclave.name} copyEnabled />
        </GridItem>
        <GridItem>
          <ValueCard title={"UUID"} value={enclave.shortenedUuid} copyEnabled />
        </GridItem>
        <GridItem>
          <ValueCard title={"Status"} value={<EnclaveStatus status={enclave.containersStatus} variant={"asText"} />} />
        </GridItem>
        <GridItem>
          <ValueCard
            title={"Creation Date"}
            value={
              <FormatDateTime
                dateTime={enclaveCreationDateTime}
                format={{
                  ...DateTime.TIME_24_SIMPLE,
                  weekday: "long",
                }}
              />
            }
          />
        </GridItem>
      </Grid>
      <TitledBox title={"Services"}>
        {!isDefined(enclave.services) && <Spinner />}
        {isDefined(enclave.services) && enclave.services.isOk && (
          <ServicesTable
            servicesResponse={enclave.services.value}
            enclaveUUID={enclave.enclaveUuid}
            enclaveShortUUID={enclave.shortenedUuid}
          />
        )}
        {isDefined(enclave.services) && enclave.services.isErr && <KurtosisAlert message={enclave.services.error} />}
      </TitledBox>
      <TitledBox title={"Files Artifacts"}>
        {!isDefined(enclave.filesAndArtifacts) && <Spinner />}
        {isDefined(enclave.filesAndArtifacts) && enclave.filesAndArtifacts.isOk && (
          <FilesTable filesAndArtifacts={enclave.filesAndArtifacts.value} enclave={enclave} />
        )}
        {isDefined(enclave.filesAndArtifacts) && enclave.filesAndArtifacts.isErr && (
          <KurtosisAlert message={enclave.filesAndArtifacts.error} />
        )}
      </TitledBox>
    </Flex>
  );
};

import { Flex, Grid, GridItem } from "@chakra-ui/react";
import { DateTime } from "luxon";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { isDefined } from "../../utils";
import { FormatDateTime } from "../FormatDateTime";
import { FLEX_STANDARD_GAP } from "../theme/constants";
import { TitledCard } from "../TitledCard";
import { ValueCard } from "../ValueCard";
import { FilesTable } from "./tables/FilesTable";
import { ServicesTable } from "./tables/ServicesTable";
import { EnclaveStatus } from "./widgets/EnclaveStatus";

type EnclaveOverviewProps = {
  enclave: EnclaveFullInfo;
};

export const EnclaveOverview = ({ enclave }: EnclaveOverviewProps) => {
  const enclaveCreationDateTime = isDefined(enclave.creationTime)
    ? DateTime.fromJSDate(enclave.creationTime.toDate())
    : null;

  return (
    <Flex flexDirection={"column"} gap={FLEX_STANDARD_GAP}>
      <Grid templateColumns={"repeat(4, 1fr)"} gap={FLEX_STANDARD_GAP}>
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
      <TitledCard title={"Services"}>
        <ServicesTable enclave={enclave} />
      </TitledCard>
      <TitledCard title={"Files"}>
        <FilesTable enclave={enclave} />
      </TitledCard>
    </Flex>
  );
};

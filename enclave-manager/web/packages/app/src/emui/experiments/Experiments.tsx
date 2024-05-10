import { Alert, AlertDescription, AlertIcon, Flex } from "@chakra-ui/react";
import { AppPageLayout, registerBreadcrumbHandler } from "kurtosis-ui-components";
import { ExperimentKey, useExperiments } from "./ExperimentsContext";

registerBreadcrumbHandler("experiments", () => <></>);

const Experiments = () => {
  const { experiments, toggleExperiment } = useExperiments();
  return (
    <AppPageLayout>
      <Flex pl={"6px"} p={"16px"} alignItems={"center"} justifyContent={"space-between"}>
        <Alert status={"error"}>
          <AlertIcon />
          <AlertDescription width={"100%"}>
            You have reached a secret page. Please be careful as the options exposed here may break your application.
          </AlertDescription>
        </Alert>
      </Flex>
      <Flex pl={"6px"} pb={"16px"}>
        {Object.keys(experiments).map((experiment) => (
          <Flex gap={2}>
            <input
              type="checkbox"
              name={experiment}
              checked={experiments[experiment as ExperimentKey] as boolean}
              onChange={() => toggleExperiment(experiment as ExperimentKey)}
            />
            <label htmlFor={experiment}>{experiment}</label>
          </Flex>
        ))}
      </Flex>
    </AppPageLayout>
  );
};

export default Experiments;

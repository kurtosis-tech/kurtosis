import {
  Accordion,
  AccordionButton,
  AccordionIcon,
  AccordionItem,
  AccordionPanel,
  Alert,
  AlertDescription,
  AlertIcon,
  AlertProps,
  AlertTitle,
  Box,
  Flex,
} from "@chakra-ui/react";
import { FallbackProps } from "react-error-boundary";
import { isDefined, stringifyError } from "./utils";

type KurtosisAlertProps = AlertProps & {
  message: string;
  details?: string;
};

export const KurtosisAlert = ({ message, details, ...alertProps }: KurtosisAlertProps) => {
  return (
    <Alert status="error" overflowY={"auto"} maxHeight={"300px"} alignItems={"flex-start"} {...alertProps}>
      <AlertIcon />
      <Flex flexDirection={"column"} width={"100%"} gap={"8px"}>
        <Flex direction={"row"}>
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{message}</AlertDescription>
        </Flex>
        {isDefined(details) && (
          <Accordion allowToggle>
            <AccordionItem>
              <h2>
                <AccordionButton>
                  <Box as="span" flex="1" textAlign="left">
                    Error details
                  </Box>
                  <AccordionIcon />
                </AccordionButton>
              </h2>
              <AccordionPanel pb={4}>
                <Box as={"pre"} whiteSpace={"pre-wrap"} wordBreak={"break-word"}>
                  {details}
                </Box>
              </AccordionPanel>
            </AccordionItem>
          </Accordion>
        )}
      </Flex>
    </Alert>
  );
};

export const KurtosisAlertError = ({ error }: FallbackProps) => {
  return <KurtosisAlert message={"An error ocurred, details below"} details={stringifyError(error)} />;
};

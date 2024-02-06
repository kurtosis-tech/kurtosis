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
  CloseButton,
  Flex,
} from "@chakra-ui/react";
import { FallbackProps } from "react-error-boundary";
import { isDefined, stringifyError } from "./utils";

type KurtosisAlertProps = AlertProps & {
  message: string;
  details?: string;
  onClose?: () => void;
};

export const KurtosisAlert = ({ message, details, onClose, ...alertProps }: KurtosisAlertProps) => {
  return (
    <Alert
      status="error"
      overflowY={"auto"}
      maxHeight={"300px"}
      position="relative"
      alignItems={"flex-start"}
      {...alertProps}
    >
      <AlertIcon />
      {isDefined(onClose) && (
        <CloseButton alignSelf="flex-start" position="absolute" right={"4px"} top={"4px"} onClick={onClose} />
      )}
      <Flex flexDirection={"column"} width={"100%"} gap={"8px"}>
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
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

import { CheckCircleIcon } from "@chakra-ui/icons";
import { Flex, forwardRef, Icon, Text } from "@chakra-ui/react";

type ToastProps = {
  message: string;
};

export const SuccessToast = forwardRef<ToastProps, "div">(({ message }: ToastProps, ref) => {
  return (
    <Flex ref={ref} bg={"rgba(0, 194, 35, 0.24)"} p={"6px 16px"} borderRadius={"6px"} gap={"12px"}>
      <Icon height={"24px"} width={"24px"} as={CheckCircleIcon} color={"kurtosisGreen.400"} />
      <Text fontWeight={"bold"} fontSize={"lg"}>
        {message}
      </Text>
    </Flex>
  );
});

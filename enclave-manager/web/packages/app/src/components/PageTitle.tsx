import { Heading, HeadingProps } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

type PageTitleProps = PropsWithChildren<HeadingProps>;

export const PageTitle = ({ children, ...headingProps }: PageTitleProps) => {
  return (
    <Heading fontSize={"lg"} fontWeight={"medium"} pl={"8px"} {...headingProps}>
      {children}
    </Heading>
  );
};

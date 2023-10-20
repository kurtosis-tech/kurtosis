import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbSeparator } from "@chakra-ui/react";
import { Link } from "react-router-dom";
import { ChevronRightIcon } from "@chakra-ui/icons";

export type EnclaveBreadCrumb = {
  name: string;
  destination: string;
};

type EnclaveBreadcrumbsProps = {
  crumbs: EnclaveBreadCrumb[];
};

export const Breadcrumbs = ({ crumbs }: EnclaveBreadcrumbsProps) => {
  return (
    <Breadcrumb variant={"topNavigation"} separator={<ChevronRightIcon h={"24px"} />}>
      {crumbs.map(({ name, destination }, i, arr) => (
        <BreadcrumbItem key={i} isCurrentPage={i === arr.length - 1}>
          <BreadcrumbLink as={i === arr.length - 1 ? undefined : Link}>{name}</BreadcrumbLink>
        </BreadcrumbItem>
      ))}
    </Breadcrumb>
  );
};

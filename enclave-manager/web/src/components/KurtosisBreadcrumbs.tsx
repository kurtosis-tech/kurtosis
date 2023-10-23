import { Breadcrumb, BreadcrumbItem, BreadcrumbLink } from "@chakra-ui/react";
import { Link } from "react-router-dom";
import { ChevronRightIcon } from "@chakra-ui/icons";

export type KurtosisBreadcrumb = {
  name: string;
  destination: string;
};

type KurtosisBreadcrumbsProps = {
  crumbs: KurtosisBreadcrumb[];
};

export const KurtosisBreadcrumbs = ({ crumbs }: KurtosisBreadcrumbsProps) => {
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

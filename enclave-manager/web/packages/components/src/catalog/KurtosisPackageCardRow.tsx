import { Button, ButtonGroup, Flex, IconButton, Text } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { memo, ReactElement, useEffect, useRef, useState } from "react";
import { MdArrowBackIosNew, MdArrowForwardIos } from "react-icons/md";
import { isDefined } from "../utils";
import { KurtosisPackageCard } from "./KurtosisPackageCard";

type KurtosisPackageCardRowProps = {
  title: string;
  icon?: ReactElement;
  packages: KurtosisPackage[];
  onSeeAllClicked?: () => void;
  onPackageRunClicked: (kurtosisPackage: KurtosisPackage) => void;
};

export const KurtosisPackageCardRow = memo(
  ({ packages, onSeeAllClicked, onPackageRunClicked, title, icon }: KurtosisPackageCardRowProps) => {
    const flexRef = useRef<HTMLDivElement>(null);
    const [scrollPosition, setScrollPosition] = useState<"start" | "mid" | "end" | "not-scrollable">("start");

    const handleScrollLeft = () => {
      if (isDefined(flexRef.current)) {
        flexRef.current.scrollBy({ left: -200, top: 0, behavior: "smooth" });
      }
    };
    const handleScrollRight = () => {
      if (isDefined(flexRef.current)) {
        flexRef.current.scrollBy({ left: 200, top: 0, behavior: "smooth" });
      }
    };

    useEffect(() => {
      if (isDefined(flexRef.current)) {
        const updateScrollPosition = () => {
          if (flexRef.current) {
            if (flexRef.current.scrollWidth === flexRef.current.clientWidth) {
              setScrollPosition("not-scrollable");
            } else if (flexRef.current.scrollLeft <= 0) {
              setScrollPosition("start");
            } else if (flexRef.current.scrollLeft >= flexRef.current.scrollWidth - flexRef.current.clientWidth) {
              setScrollPosition("end");
            } else {
              setScrollPosition("mid");
            }
          }
        };

        window.addEventListener("resize", updateScrollPosition);
        flexRef.current.addEventListener("scroll", updateScrollPosition);
        return () => {
          if (isDefined(flexRef.current)) {
            window.removeEventListener("resize", updateScrollPosition);
            flexRef.current.removeEventListener("scroll", updateScrollPosition);
          }
        };
      }
    }, [flexRef.current]);

    return (
      <Flex flexDirection={"column"}>
        <Flex justifyContent={"space-between"} fontSize={"lg"} pb={"16px"} fontWeight={"medium"}>
          <Flex gap={"8px"} alignItems={"center"}>
            {icon}
            <Text as={"span"}>{title}</Text>
          </Flex>
          <Flex gap={"16px"}>
            {isDefined(onSeeAllClicked) && (
              <Button variant={"ghost"} onClick={onSeeAllClicked} size={"xs"}>
                See all
              </Button>
            )}
            <ButtonGroup isAttached variant={"ghost"} size={"xs"}>
              <IconButton
                aria-label={"Scroll left"}
                onClick={handleScrollLeft}
                icon={<MdArrowBackIosNew />}
                isDisabled={scrollPosition === "start" || scrollPosition === "not-scrollable"}
              />
              <IconButton
                aria-label={"Scroll right"}
                onClick={handleScrollRight}
                icon={<MdArrowForwardIos />}
                isDisabled={scrollPosition === "end" || scrollPosition === "not-scrollable"}
              />
            </ButtonGroup>
          </Flex>
        </Flex>
        <Flex ref={flexRef} gap={"32px"} rowGap={"32px"} overflowX={"auto"} justifyContent={"flex-start"}>
          {packages.map((kurtosisPackage) => (
            <KurtosisPackageCard
              kurtosisPackage={kurtosisPackage}
              onRunClick={() => onPackageRunClicked(kurtosisPackage)}
            />
          ))}
        </Flex>
      </Flex>
    );
  },
);

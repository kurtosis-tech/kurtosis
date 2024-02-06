import { Tab, TabList } from "@chakra-ui/react";
import { useState } from "react";
import { isDefined } from "./utils";

type HoverLineTabListProps = {
  tabs: readonly string[];
  activeTab: string;
};

/**
 * This component is needed as the hover interaction cannot be controlled with CSS
 */
export const HoverLineTabList = ({ tabs, activeTab }: HoverLineTabListProps) => {
  const [hoveredTab, setHoveredTab] = useState<string>();

  return (
    <TabList>
      {tabs.map((tab) => {
        return (
          <Tab
            key={tab}
            sx={{
              _selected: {
                borderBottom:
                  hoveredTab === tab || (!isDefined(hoveredTab) && activeTab === tab)
                    ? "2px solid var(--chakra-colors-kurtosisGreen-400)"
                    : undefined,
              },
              borderBottom:
                hoveredTab === tab || (!isDefined(hoveredTab) && activeTab === tab)
                  ? "2px solid var(--chakra-colors-kurtosisGreen-400)"
                  : "2px solid transparent",
            }}
            onMouseEnter={() => setHoveredTab(tab)}
            onMouseLeave={() => setHoveredTab(undefined)}
          >
            {tab}
          </Tab>
        );
      })}
    </TabList>
  );
};

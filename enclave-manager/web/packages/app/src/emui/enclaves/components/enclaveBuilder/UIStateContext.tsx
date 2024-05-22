import { createContext, PropsWithChildren, useContext, useState } from "react";

// Define the context
const UIContext = createContext<{
  expandedNodes: Record<string, boolean>;
  toggleExpanded: (nodeId: string) => void;
}>({
  expandedNodes: {},
  toggleExpanded: () => {},
});

// Define the provider
export const UIStateProvider = ({ children }: PropsWithChildren) => {
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});

  const toggleExpanded = (nodeId: string) => {
    setExpandedNodes({ ...expandedNodes, [nodeId]: !expandedNodes[nodeId] });
  };

  return <UIContext.Provider value={{ expandedNodes, toggleExpanded }}>{children}</UIContext.Provider>;
};

// Define the hook
export function useUIState() {
  return useContext(UIContext);
}

import { createContext, PropsWithChildren, useContext, useState } from "react";

export type ExperimentKey = "enableCloudVersionUpgrade";

interface ExperimentsContextProps {
  experiments: Record<ExperimentKey, boolean>;
  toggleExperiment: (featureName: ExperimentKey) => void;
}

export const ExperimentsContext = createContext<ExperimentsContextProps | undefined>(undefined);

// This context provider stores temporary experiment flags that can be toggled on and off
// using the /experiments page. These are only stored in react state and are not persisted.
export const ExperimentsContextProvider = ({ children }: PropsWithChildren) => {
  const [experiments, setExperiments] = useState<ExperimentsContextProps["experiments"]>({
    enableCloudVersionUpgrade: false,
  });

  const toggleExperiment = (featureName: ExperimentKey) => {
    setExperiments((prevExperiments) => ({
      ...prevExperiments,
      [featureName]: !prevExperiments[featureName],
    }));
  };

  return (
    <ExperimentsContext.Provider value={{ experiments, toggleExperiment }}>{children}</ExperimentsContext.Provider>
  );
};

export const useExperiments = () => {
  const context = useContext(ExperimentsContext);
  if (context === undefined) {
    throw new Error("useExperiments must be used within a ExperimentsContextProvider");
  }
  return context;
};

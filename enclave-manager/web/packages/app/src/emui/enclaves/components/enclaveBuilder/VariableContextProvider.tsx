import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from "react";
import { KurtosisNodeData, Variable } from "./types";
import { getVariablesFromNodes } from "./utils";

type VariableContextState = {
  data: Record<string, KurtosisNodeData>;
  variables: Variable[];
  updateData: (id: string, data: KurtosisNodeData | ((oldData: KurtosisNodeData) => KurtosisNodeData)) => void;
  removeData: (id: { id: string }[]) => void;
  initialImportedPackageData: null | Record<string, KurtosisNodeData>;
  setInitialImportedPackageData: (data: Record<string, KurtosisNodeData>) => void;
};

const VariableContext = createContext<VariableContextState>({
  data: {},
  variables: [],
  updateData: () => null,
  removeData: () => null,
  initialImportedPackageData: null,
  setInitialImportedPackageData: () => null,
});

type VariableContextProviderProps = {
  initialData: Record<string, KurtosisNodeData>;
};

export const VariableContextProvider = ({ initialData, children }: PropsWithChildren<VariableContextProviderProps>) => {
  const [data, setData] = useState<Record<string, KurtosisNodeData>>(initialData);
  // A snapshot of the imported package data so we can compare changes the user
  // makes to the original configuration and generate the appropriate Starlark code
  const [initialImportedPackageData, setInitialImportedPackageData] = useState<Record<string, KurtosisNodeData> | null>(
    null,
  );

  const variables = useMemo((): Variable[] => {
    return getVariablesFromNodes(data);
  }, [data]);

  const updateData = useCallback(
    (id: string, data: KurtosisNodeData | ((oldData: KurtosisNodeData) => KurtosisNodeData)) => {
      setData((oldData) => ({ ...oldData, [id]: typeof data === "object" ? data : data(oldData[id]) }));
    },
    [],
  );

  const removeData = useCallback((ids: { id: string }[]) => {
    setData((oldData) => {
      const r = { ...oldData };
      for (const { id } of ids) {
        delete r[id];
      }
      return r;
    });
  }, []);

  return (
    <VariableContext.Provider
      value={{ data, variables, updateData, removeData, setInitialImportedPackageData, initialImportedPackageData }}
    >
      {children}
    </VariableContext.Provider>
  );
};

export const useVariableContext = () => useContext(VariableContext);

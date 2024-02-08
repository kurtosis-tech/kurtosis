import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from "react";
import { KurtosisServiceNodeData } from "./KurtosisServiceNode";
import { Variable } from "./types";
import { getVariablesFromNodes } from "./utils";

type VariableContextState = {
  data: Record<string, KurtosisServiceNodeData>;
  variables: Variable[];
  updateData: (id: string, data: KurtosisServiceNodeData) => void;
  removeData: (id: string) => void;
};

const VariableContext = createContext<VariableContextState>({
  data: {},
  variables: [],
  updateData: () => null,
  removeData: () => null,
});

type VariableContextProviderProps = {
  initialData: Record<string, KurtosisServiceNodeData>;
};

export const VariableContextProvider = ({ initialData, children }: PropsWithChildren<VariableContextProviderProps>) => {
  const [data, setData] = useState<Record<string, KurtosisServiceNodeData>>(initialData);

  const variables = useMemo((): Variable[] => {
    return getVariablesFromNodes(data);
  }, [data]);

  const updateData = useCallback((id: string, data: KurtosisServiceNodeData) => {
    setData((oldData) => ({ ...oldData, [id]: data }));
  }, []);

  const removeData = useCallback((id: string) => {
    setData((oldData) => {
      const r = { ...oldData };
      delete r[id];
      return r;
    });
  }, []);

  return (
    <VariableContext.Provider value={{ data, variables, updateData, removeData }}>{children}</VariableContext.Provider>
  );
};

export const useVariableContext = () => useContext(VariableContext);

import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from "react";
import { KurtosisNodeData, Variable } from "./types";
import { getVariablesFromNodes } from "./utils";

type VariableContextState = {
  data: Record<string, KurtosisNodeData>;
  variables: Variable[];
  updateData: (id: string, data: KurtosisNodeData) => void;
  removeData: (id: string) => void;
};

const VariableContext = createContext<VariableContextState>({
  data: {},
  variables: [],
  updateData: () => null,
  removeData: () => null,
});

type VariableContextProviderProps = {
  initialData: Record<string, KurtosisNodeData>;
};

export const VariableContextProvider = ({ initialData, children }: PropsWithChildren<VariableContextProviderProps>) => {
  const [data, setData] = useState<Record<string, KurtosisNodeData>>(initialData);

  const variables = useMemo((): Variable[] => {
    return getVariablesFromNodes(data);
  }, [data]);

  const updateData = useCallback((id: string, data: KurtosisNodeData) => {
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

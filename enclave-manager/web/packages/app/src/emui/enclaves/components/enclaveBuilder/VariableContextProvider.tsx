import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from "react";
import { KurtosisNodeData, Variable } from "./types";
import { getVariablesFromNodes } from "./utils";

type VariableContextState = {
  data: Record<string, KurtosisNodeData>;
  variables: Variable[];
  updateData: (id: string, data: KurtosisNodeData | ((oldData: KurtosisNodeData) => KurtosisNodeData)) => void;
  removeData: (id: { id: string }[]) => void;
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

  const updateData = useCallback(
    (id: string, data: KurtosisNodeData | ((oldData: KurtosisNodeData) => KurtosisNodeData)) => {
      console.log(`${id} was updated`);
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
    <VariableContext.Provider value={{ data, variables, updateData, removeData }}>{children}</VariableContext.Provider>
  );
};

export const useVariableContext = () => useContext(VariableContext);

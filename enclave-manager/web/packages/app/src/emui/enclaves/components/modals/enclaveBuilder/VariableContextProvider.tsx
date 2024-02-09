import { createContext, PropsWithChildren, useCallback, useContext, useMemo, useState } from "react";
import { Variable } from "./types";
import { getVariablesFromNodes } from "./utils";

export type KurtosisPort = {
  portName: string;
  port: number;
  transportProtocol: "TCP" | "UDP";
  applicationProtocol: string;
};

export type KurtosisFileMount = {
  mountPoint: string;
  artifactName: string;
};

export type KurtosisServiceNodeData = {
  type: "service";
  serviceName: string;
  image: string;
  env: { key: string; value: string }[];
  ports: KurtosisPort[];
  files: KurtosisFileMount[];
  isValid: boolean;
};

export type KurtosisArtifactNodeData = {
  type: "artifact";
  artifactName: string;
  files: Record<string, string>;
  isValid: boolean;
};

export type KurtosisNodeData = KurtosisArtifactNodeData | KurtosisServiceNodeData;

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

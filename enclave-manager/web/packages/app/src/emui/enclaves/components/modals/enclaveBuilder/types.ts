export type Variable = {
  id: string;
  displayName: string;
  value: string;
};

export type KurtosisPort = {
  portName: string;
  port: number;
  transportProtocol: "TCP" | "UDP";
  applicationProtocol: string;
};

export type KurtosisEnvironmentVar = { key: string; value: string };

export type KurtosisFileMount = {
  mountPoint: string;
  artifactName: string;
};
export type KurtosisServiceNodeData = {
  type: "service";
  serviceName: string;
  image: string;
  env: KurtosisEnvironmentVar[];
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

export type KurtosisShellNodeData = {
  type: "shell";
  shellName: string;
  command: string;
  image: string;
  env: KurtosisEnvironmentVar[];
  files: KurtosisFileMount[];
  store: { value: string }[];
  wait_enabled: "true" | "false";
  wait: string;
  isValid: boolean;
};

export type KurtosisNodeData = KurtosisArtifactNodeData | KurtosisServiceNodeData | KurtosisShellNodeData;

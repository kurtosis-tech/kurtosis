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

export type KurtosisAcceptableCode = {
  value: number;
};

export type KurtosisImageType = "image" | "dockerfile" | "nix";

export type KurtosisImageConfig = {
  image: string;
  type: KurtosisImageType;
  registryUsername: string;
  registryPassword: string;
  registry: string;
  buildContextDir: string;
  targetStage: string;
  flakeLocationDir: string;
  flakeOutput: string;
};

export type KurtosisServiceNodeData = {
  type: "service";
  serviceName: string;
  image: KurtosisImageConfig;
  env: KurtosisEnvironmentVar[];
  ports: KurtosisPort[];
  files: KurtosisFileMount[];
  execStepEnabled: "true" | "false";
  execStepCommand: string;
  execStepAcceptableCodes: KurtosisAcceptableCode[];
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
  image: KurtosisImageConfig;
  env: KurtosisEnvironmentVar[];
  files: KurtosisFileMount[];
  store: string;
  wait_enabled: "true" | "false";
  wait: string;
  isValid: boolean;
};

export type KurtosisPythonPackage = { packageName: string };
export type KurtosisPythonArg = { arg: string };

export type KurtosisPythonNodeData = {
  type: "python";
  pythonName: string;
  command: string;
  image: KurtosisImageConfig;
  packages: KurtosisPythonPackage[];
  args: KurtosisPythonArg[];
  files: KurtosisFileMount[];
  store: string;
  wait_enabled: "true" | "false";
  wait: string;
  isValid: boolean;
};

export type KurtosisPackageNodeData = {
  type: "package";
  name: string;
  packageId: string;
  args: Record<string, any>;
  isValid: boolean;
};

export type KurtosisNodeData =
  | KurtosisArtifactNodeData
  | KurtosisServiceNodeData
  | KurtosisShellNodeData
  | KurtosisPythonNodeData
  | KurtosisPackageNodeData;

export type PlanPort = {
  name: string;
  number: number;
  transportProtocol: "TCP" | "UDP";
  applicationProtocol: string;
};

export type PlanService = {
  name: string;
  uuid: string;
  image: string;
  envVars: KurtosisEnvironmentVar[];
  ports: PlanPort[];
};

export type PlanYaml = {
  packageId: string;
  services: PlanService[];
};

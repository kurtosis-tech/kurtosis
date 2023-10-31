export type ConfigureEnclaveForm = {
  enclaveName: string;
  restartServices: boolean;
  args: Record<string, any>;
};

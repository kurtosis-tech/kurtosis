import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { CSSProperties, forwardRef, PropsWithChildren, useImperativeHandle } from "react";
import { FormProvider, SubmitHandler, useForm, useFormContext } from "react-hook-form";
import { ConfigureEnclaveForm } from "./types";
import { transformFormArgsToKurtosisArgs } from "./utils";

type EnclaveConfigurationFormProps = PropsWithChildren<{
  onSubmit: SubmitHandler<ConfigureEnclaveForm>;
  kurtosisPackage: KurtosisPackage;
  initialValues?: ConfigureEnclaveForm;
  style?: CSSProperties;
}>;

export type EnclaveConfigurationFormImperativeAttributes = {
  getValues: () => ConfigureEnclaveForm;
  setValues: (key: keyof ConfigureEnclaveForm, value: any) => void;
  isDirty: () => boolean;
};

export const EnclaveConfigurationForm = forwardRef<
  EnclaveConfigurationFormImperativeAttributes,
  EnclaveConfigurationFormProps
>(({ children, kurtosisPackage, onSubmit, initialValues, style }: EnclaveConfigurationFormProps, ref) => {
  const methods = useForm<ConfigureEnclaveForm>({ values: initialValues });

  useImperativeHandle(
    ref,
    () => ({
      getValues: () => {
        return methods.getValues();
      },
      setValues: (key: keyof ConfigureEnclaveForm, value: any) => {
        methods.setValue(key, value);
      },
      isDirty: () => {
        if (isDefined(initialValues)) {
          return Object.keys(methods.formState.dirtyFields.args || {}).length > 0;
        } else {
          return Object.keys(methods.getValues().args || {}).length > 0;
        }
      },
    }),
    [methods, initialValues],
  );

  const handleSubmit: SubmitHandler<ConfigureEnclaveForm> = (data: { args: { [x: string]: any } }) => {
    onSubmit({
      enclaveName: "",
      restartServices: false,
      ...data,
      args: transformFormArgsToKurtosisArgs(data.args, kurtosisPackage),
    });
  };

  return (
    <FormProvider {...methods}>
      <form style={style} onSubmit={methods.handleSubmit(handleSubmit)}>
        {children}
      </form>
    </FormProvider>
  );
});

export const useEnclaveConfigurationFormContext = () => useFormContext<ConfigureEnclaveForm>();

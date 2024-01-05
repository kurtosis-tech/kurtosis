import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { CSSProperties, forwardRef, PropsWithChildren, useEffect, useImperativeHandle, useState } from "react";
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
  const [isDirty, setIsDirty] = useState(false);
  const methods = useForm<ConfigureEnclaveForm>({ defaultValues: initialValues });

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
        return isDirty;
      },
    }),
    [methods, initialValues, isDirty],
  );

  useEffect(() => {
    const { unsubscribe } = methods.watch((value) => {
      // We manually track modified fields because we dynamically register values (this means that the
      // isDirty field on the react-hook-form state cannot be used). Relying on the react-hook-form state isDirty field
      // will actually trigger a cyclic setState (as it's secretly a getter method).
      setIsDirty(true);
    });
    return () => unsubscribe();
  }, [methods]);

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
EnclaveConfigurationForm.displayName = "EnclaveConfigurationForm";

export const useEnclaveConfigurationFormContext = () => useFormContext<ConfigureEnclaveForm>();

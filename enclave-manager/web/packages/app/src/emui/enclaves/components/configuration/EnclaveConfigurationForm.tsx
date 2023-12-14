import { ArgumentValueType, KurtosisPackage, PackageArg } from "kurtosis-cloud-indexer-sdk";
import { isDefined, isStringTrue } from "kurtosis-ui-components";
import { CSSProperties, forwardRef, PropsWithChildren, useImperativeHandle } from "react";
import { FormProvider, SubmitHandler, useForm, useFormContext } from "react-hook-form";
import YAML from "yaml";
import { ConfigureEnclaveForm } from "./types";

type EnclaveConfigurationFormProps = PropsWithChildren<{
  onSubmit: SubmitHandler<ConfigureEnclaveForm>;
  kurtosisPackage: KurtosisPackage;
  initialValues?: ConfigureEnclaveForm;
  style?: CSSProperties;
}>;

export type EnclaveConfigurationFormImperativeAttributes = {
  getValues: () => ConfigureEnclaveForm;
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
    }),
    [methods],
  );

  const handleSubmit: SubmitHandler<ConfigureEnclaveForm> = (data: { args: { [x: string]: any } }) => {
    const transformValue = (
      valueType: ArgumentValueType | undefined,
      value: any,
      innerValuetype?: ArgumentValueType,
    ) => {
      // The DICT type is stored as an array of {key, value} objects, before passing it up we should correct
      // any instances of it to be Record<string, any> objects
      const transformRecordsToObject = (records: { key: string; value: any }[], valueType?: ArgumentValueType) =>
        records.reduce(
          (acc, { key, value }) => ({
            ...acc,
            [key]: valueType === ArgumentValueType.BOOL ? isStringTrue(value) : value,
          }),
          {},
        );

      switch (valueType) {
        case ArgumentValueType.DICT:
          if (!isDefined(value)) return {};
          else return transformRecordsToObject(value, innerValuetype);
        case ArgumentValueType.LIST:
          return value.map((v: any) => transformValue(innerValuetype, v));
        case ArgumentValueType.BOOL:
          return isDefined(value) ? isStringTrue(value) : null;
        case ArgumentValueType.INTEGER:
          return isNaN(value) || isNaN(parseFloat(value)) ? null : parseFloat(value);
        case ArgumentValueType.STRING:
          return value;
        case ArgumentValueType.JSON:
          return YAML.parse(value);
        default:
          return value;
      }
    };

    const newArgs: Record<string, any> = kurtosisPackage.args
      .filter((arg) => arg.name !== "plan") // plan args needs to be filtered out as it's not an actual arg
      .map((arg): [PackageArg, any] => [
        arg,
        transformValue(
          arg.typeV2?.topLevelType,
          data.args[arg.name],
          arg.typeV2?.topLevelType === ArgumentValueType.LIST ? arg.typeV2?.innerType1 : arg.typeV2?.innerType2,
        ),
      ])
      .filter(([arg, value]) => {
        switch (arg.typeV2?.topLevelType) {
          case ArgumentValueType.DICT:
            return Object.keys(value).length > 0;
          case ArgumentValueType.LIST:
            return value.length > 0;
          case ArgumentValueType.STRING:
            return isDefined(value) && value.length > 0;
          default:
            return isDefined(value);
        }
      })
      .reduce(
        (acc, [arg, value]) => ({
          ...acc,
          [arg.name]: value,
        }),
        {},
      );

    onSubmit({ enclaveName: "", restartServices: false, ...data, args: newArgs });
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

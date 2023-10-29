import { PropsWithChildren } from "react";
import { FormProvider, SubmitHandler, useForm, useFormContext } from "react-hook-form";
import { ArgumentValueType, KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isStringTrue } from "../../../utils";
import { ConfigureEnclaveForm } from "./types";

type PackageConfigurationFormProps = PropsWithChildren<{
  onSubmit: SubmitHandler<ConfigureEnclaveForm>;
  kurtosisPackage: KurtosisPackage;
}>;

export const EnclaveConfigurationForm = ({ children, kurtosisPackage, onSubmit }: PackageConfigurationFormProps) => {
  const methods = useForm<ConfigureEnclaveForm>();

  const handleSubmit: SubmitHandler<ConfigureEnclaveForm> = (data) => {
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

    const newArgs: Record<string, any> = kurtosisPackage.args.reduce(
      (acc, cur) => ({
        ...acc,
        [cur.name]:
          cur.typeV2?.topLevelType === ArgumentValueType.DICT
            ? transformRecordsToObject(data.args[cur.name], cur.typeV2.innerType2)
            : cur.typeV2?.topLevelType === ArgumentValueType.BOOL
            ? isStringTrue(data.args[cur.name])
            : data.args[cur.name],
      }),
      {},
    );
    onSubmit({ ...data, args: newArgs });
  };

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(handleSubmit)}>{children}</form>
    </FormProvider>
  );
};

export const useEnclaveConfigurationFormContext = () => useFormContext<ConfigureEnclaveForm>();

import { ArgumentValueType, PackageArg } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { KurtosisArgumentTypeInput } from "./inputs/KurtosisArgumentTypeInput";
import { KurtosisArgumentFormControl } from "./KurtosisArgumentFormControl";
import { argToTypeString } from "./utils";

type KurtosisPackageArgumentInputProps = {
  argument: PackageArg;
  disabled?: boolean;
};

export const KurtosisPackageArgumentInput = ({ argument, disabled }: KurtosisPackageArgumentInputProps) => {
  if (argument.name === "plan") {
    // The 'plan' argument is internal and is not used.
    return null;
  }

  const fieldName: `args.${string}` = `args.${argument.name}`;

  return (
    <KurtosisArgumentFormControl
      name={fieldName}
      label={argument.name}
      type={argToTypeString(argument)}
      disabled={disabled}
      isRequired={argument.isRequired}
      helperText={argument.description}
    >
      <KurtosisArgumentTypeInput
        type={argument.typeV2?.topLevelType || ArgumentValueType.JSON}
        subType1={argument.typeV2?.innerType1}
        subType2={argument.typeV2?.innerType2}
        name={fieldName}
        isRequired={argument.isRequired}
      />
    </KurtosisArgumentFormControl>
  );
};

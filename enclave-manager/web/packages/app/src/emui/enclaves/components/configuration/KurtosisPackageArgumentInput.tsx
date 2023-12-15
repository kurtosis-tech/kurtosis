import { PackageArg } from "kurtosis-cloud-indexer-sdk";
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
  const label = argument.name
    .split("_")
    .filter((w) => w.length > 0)
    .map((w) => `${w[0].toUpperCase()}${w.substring(1)}`)
    .join(" ");

  return (
    <KurtosisArgumentFormControl
      name={fieldName}
      label={label}
      type={argToTypeString(argument)}
      disabled={disabled}
      isRequired={argument.isRequired}
      helperText={argument.description}
    >
      <KurtosisArgumentTypeInput
        type={argument.typeV2?.topLevelType}
        subType1={argument.typeV2?.innerType1}
        subType2={argument.typeV2?.innerType2}
        name={fieldName}
        placeholder={argument.defaultValue}
        isRequired={argument.isRequired}
      />
    </KurtosisArgumentFormControl>
  );
};

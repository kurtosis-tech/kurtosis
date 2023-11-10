import { FieldPath } from "react-hook-form";
import { ArgumentValueType } from "../../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { assertDefined } from "../../../../utils";
import { ConfigureEnclaveForm } from "../types";
import { BooleanArgumentInput } from "./BooleanArgumentInput";
import { DictArgumentInput } from "./DictArgumentInput";
import { IntegerArgumentInput } from "./IntegerArgumentInput";
import { JSONArgumentInput } from "./JSONArgumentInput";
import { ListArgumentInput } from "./ListArgumentInput";
import { StringArgumentInput } from "./StringArgumentInput";

export type KurtosisArgumentTypeInputProps = {
  type?: ArgumentValueType;
  subType1?: ArgumentValueType;
  subType2?: ArgumentValueType;
  name: FieldPath<ConfigureEnclaveForm>;
  isRequired?: boolean;
  validate?: (value: any) => string | undefined;
  disabled?: boolean;
};

export const KurtosisArgumentTypeInput = ({
  type,
  subType1,
  subType2,
  name,
  isRequired,
  validate,
  disabled,
}: KurtosisArgumentTypeInputProps) => {
  switch (type) {
    case ArgumentValueType.INTEGER:
      return <IntegerArgumentInput name={name} isRequired={isRequired} disabled={disabled} validate={validate} />;
    case ArgumentValueType.DICT:
      assertDefined(subType1, `innerType1 was not defined on DICT argument ${name}`);
      assertDefined(subType2, `innerType2 was not defined on DICT argument ${name}`);
      return (
        <DictArgumentInput
          name={name}
          isRequired={isRequired}
          keyType={subType1}
          valueType={subType2}
          validate={validate}
          disabled={disabled}
        />
      );
    case ArgumentValueType.LIST:
      assertDefined(subType1, `innerType1 was not defined on DICT argument ${name}`);
      return (
        <ListArgumentInput
          name={name}
          isRequired={isRequired}
          valueType={subType1}
          validate={validate}
          disabled={disabled}
        />
      );
    case ArgumentValueType.BOOL:
      return <BooleanArgumentInput name={name} isRequired={isRequired} validate={validate} disabled={disabled} />;
    case ArgumentValueType.STRING:
      return <StringArgumentInput name={name} isRequired={isRequired} validate={validate} disabled={disabled} />;
    case ArgumentValueType.JSON:
    default:
      return <JSONArgumentInput name={name} isRequired={isRequired} validate={validate} disabled={disabled} />;
  }
};

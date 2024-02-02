import { ArgumentValueType } from "kurtosis-cloud-indexer-sdk";
import { assertDefined } from "kurtosis-ui-components";
import { BooleanArgumentInput } from "../form/BooleanArgumentInput";
import { DictArgumentInput } from "../form/DictArgumentInput";
import { IntegerArgumentInput } from "../form/IntegerArgumentInput";
import { JSONArgumentInput } from "../form/JSONArgumentInput";
import { ListArgumentInput } from "../form/ListArgumentInput";
import { StringArgumentInput } from "../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../form/types";
import { ConfigureEnclaveForm } from "./types";

type KurtosisArgumentTypeInputProps = KurtosisFormInputProps<ConfigureEnclaveForm> & {
  type?: ArgumentValueType;
  subType1?: ArgumentValueType;
  subType2?: ArgumentValueType;
};

export const KurtosisArgumentTypeInput = ({
  type,
  subType1,
  subType2,
  name,
  placeholder,
  isRequired,
  validate,
  disabled,
  width,
  size,
  tabIndex,
}: KurtosisArgumentTypeInputProps) => {
  const childProps: KurtosisFormInputProps<ConfigureEnclaveForm> = {
    name,
    placeholder,
    isRequired,
    validate,
    disabled,
    width,
    size,
    tabIndex,
  };

  switch (type) {
    case ArgumentValueType.INTEGER:
      return <IntegerArgumentInput {...childProps} />;
    case ArgumentValueType.DICT:
      assertDefined(
        subType1,
        `innerType1 was not defined on DICT argument ${name}, check the format used matches https://docs.kurtosis.com/api-reference/starlark-reference/docstring-syntax#types`,
      );
      assertDefined(
        subType2,
        `innerType2 was not defined on DICT argument ${name}, check the format used matches https://docs.kurtosis.com/api-reference/starlark-reference/docstring-syntax#types`,
      );
      return (
        <DictArgumentInput
          renderKeyFieldInput={(props) => (
            <KurtosisArgumentTypeInput type={subType1} {...props} name={props.name as `args.${string}.${number}.key`} />
          )}
          renderValueFieldInput={(props) => (
            <KurtosisArgumentTypeInput
              type={subType2}
              {...props}
              name={props.name as `args.${string}.${number}.value`}
            />
          )}
          {...childProps}
        />
      );
    case ArgumentValueType.LIST:
      assertDefined(
        subType1,
        `innerType1 was not defined on DICT argument ${name}, check the format used matches https://docs.kurtosis.com/api-reference/starlark-reference/docstring-syntax#types`,
      );
      return (
        <ListArgumentInput
          renderFieldInput={(props) => (
            <KurtosisArgumentTypeInput
              type={subType1}
              {...props}
              name={`${props.name}.value` as `args.${string}.${number}.value`}
              width={"411px"}
              size={"sm"}
            />
          )}
          createNewValue={() => ({ value: "" })}
          {...childProps}
        />
      );
    case ArgumentValueType.BOOL:
      return <BooleanArgumentInput {...childProps} />;
    case ArgumentValueType.STRING:
      return <StringArgumentInput {...childProps} type={undefined} />;
    case ArgumentValueType.JSON:
    default:
      return <JSONArgumentInput {...childProps} />;
  }
};

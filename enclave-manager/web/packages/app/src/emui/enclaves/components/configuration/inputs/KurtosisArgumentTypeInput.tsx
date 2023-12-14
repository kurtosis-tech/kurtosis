import * as CSS from "csstype";
import { ArgumentValueType } from "kurtosis-cloud-indexer-sdk";
import { assertDefined } from "kurtosis-ui-components";
import { FieldPath } from "react-hook-form";
import { ConfigureEnclaveForm } from "../types";
import { BooleanArgumentInput } from "./BooleanArgumentInput";
import { DictArgumentInput } from "./DictArgumentInput";
import { IntegerArgumentInput } from "./IntegerArgumentInput";
import { JSONArgumentInput } from "./JSONArgumentInput";
import { ListArgumentInput } from "./ListArgumentInput";
import { StringArgumentInput } from "./StringArgumentInput";

type KurtosisArgumentTypeInputProps = {
  type?: ArgumentValueType;
  subType1?: ArgumentValueType;
  subType2?: ArgumentValueType;
  name: FieldPath<ConfigureEnclaveForm>;
  placeholder?: string;
  isRequired?: boolean;
  validate?: (value: any) => string | undefined;
  disabled?: boolean;
  width?: CSS.Property.Width;
  size?: string;
  tabIndex?: number;
};

export type KurtosisArgumentTypeInputImplProps = Omit<KurtosisArgumentTypeInputProps, "type" | "subType1" | "subType2">;

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
  const childProps: KurtosisArgumentTypeInputImplProps = {
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
      return <DictArgumentInput keyType={subType1} valueType={subType2} {...childProps} />;
    case ArgumentValueType.LIST:
      assertDefined(
        subType1,
        `innerType1 was not defined on DICT argument ${name}, check the format used matches https://docs.kurtosis.com/api-reference/starlark-reference/docstring-syntax#types`,
      );
      return <ListArgumentInput valueType={subType1} {...childProps} />;
    case ArgumentValueType.BOOL:
      return <BooleanArgumentInput {...childProps} />;
    case ArgumentValueType.STRING:
      return <StringArgumentInput {...childProps} />;
    case ArgumentValueType.JSON:
    default:
      return <JSONArgumentInput {...childProps} />;
  }
};

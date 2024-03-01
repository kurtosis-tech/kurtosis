import { ArgumentValueType, KurtosisPackage, PackageArg } from "kurtosis-cloud-indexer-sdk";
import { assertDefined, isDefined, isStringTrue } from "kurtosis-ui-components";
import YAML from "yaml";

export function argTypeToString(argType?: ArgumentValueType) {
  switch (argType) {
    case ArgumentValueType.BOOL:
      return "boolean";
    case ArgumentValueType.DICT:
      return "dictionary";
    case ArgumentValueType.INTEGER:
      return "integer";
    case ArgumentValueType.JSON:
      return "json/yaml";
    case ArgumentValueType.LIST:
      return "list";
    case ArgumentValueType.STRING:
      return "text";
    default:
      return "json/yaml";
  }
}

export function argToTypeString(arg: PackageArg) {
  switch (arg.typeV2?.topLevelType) {
    case ArgumentValueType.BOOL:
    case ArgumentValueType.STRING:
    case ArgumentValueType.INTEGER:
    case ArgumentValueType.JSON:
      return argTypeToString(arg.typeV2.topLevelType);
    case ArgumentValueType.DICT:
      return `${argTypeToString(arg.typeV2.innerType1)} -> ${argTypeToString(arg.typeV2.innerType2)}`;
    case ArgumentValueType.LIST:
      return `${argTypeToString(arg.typeV2.innerType1)} list`;
    default:
      return "json/yaml";
  }
}

export function transformFormArgsToKurtosisArgs(data: Record<string, any>, kurtosisPackage: KurtosisPackage) {
  const transformValue = (valueType: ArgumentValueType | undefined, value: any, innerValuetype?: ArgumentValueType) => {
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
        return isDefined(value) && value !== "" ? isStringTrue(value) : null;
      case ArgumentValueType.INTEGER:
        return isNaN(value) || isNaN(parseFloat(value)) ? null : parseFloat(value);
      case ArgumentValueType.STRING:
        return value;
      case ArgumentValueType.JSON:
      default:
        try {
          return YAML.parse(value);
        } catch (error: any) {
          return value;
        }
    }
  };

  const newArgs: Record<string, any> = kurtosisPackage.args
    .filter((arg) => arg.name !== "plan") // plan args needs to be filtered out as it's not an actual arg
    .map((arg): [PackageArg, any] => [
      arg,
      transformValue(
        arg.typeV2?.topLevelType,
        data[arg.name],
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
  return newArgs;
}

export function transformKurtosisArgsToFormArgs(
  data: Record<string, any>,
  kurtosisPackage: KurtosisPackage,
): Record<string, any> {
  const convertArgValue = (
    argType: ArgumentValueType | undefined,
    value: any,
    innerType1?: ArgumentValueType,
    innerType2?: ArgumentValueType,
  ): any => {
    switch (argType) {
      case ArgumentValueType.BOOL:
        return !!value ? "true" : isDefined(value) ? "false" : "";
      case ArgumentValueType.INTEGER:
        return isDefined(value) ? `${value}` : "";
      case ArgumentValueType.STRING:
        return value || "";
      case ArgumentValueType.LIST:
        assertDefined(innerType1, `Cannot parse a list argument type without knowing innerType1`);
        return isDefined(value) ? value.map((v: any) => convertArgValue(innerType1, v)) : [];
      case ArgumentValueType.DICT:
        assertDefined(innerType2, `Cannot parse a dict argument type without knowing innterType2`);
        return isDefined(value)
          ? Object.entries(value).map(([k, v]) => ({ key: k, value: convertArgValue(innerType2, v) }), {})
          : [];
      case ArgumentValueType.JSON:
      default:
        // By default, a typeless parameter is JSON.
        return isDefined(value) ? JSON.stringify(value) : "{}";
    }
  };

  const args = kurtosisPackage.args.reduce(
    (acc, arg) => ({
      ...acc,
      [arg.name]: convertArgValue(
        arg.typeV2?.topLevelType,
        data[arg.name],
        arg.typeV2?.innerType1,
        arg.typeV2?.innerType2,
      ),
    }),
    {},
  );
  return args;
}

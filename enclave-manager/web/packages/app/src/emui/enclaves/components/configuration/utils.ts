import { ArgumentValueType, PackageArg } from "kurtosis-cloud-indexer-sdk";

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

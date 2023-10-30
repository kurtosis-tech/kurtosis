import { ArgumentValueType, PackageArg } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";

export function argTypeToString(argType?: ArgumentValueType) {
  switch (argType) {
    case ArgumentValueType.BOOL:
      return "boolean";
    case ArgumentValueType.DICT:
      return "dictionary";
    case ArgumentValueType.INTEGER:
      return "integer";
    case ArgumentValueType.JSON:
      return "json";
    case ArgumentValueType.LIST:
      return "list";
    case ArgumentValueType.STRING:
      return "string";
    default:
      return "unknown";
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
      return `${argTypeToString(arg.typeV2.innerType1)}[]`;
    default:
      return "unknown";
  }
}

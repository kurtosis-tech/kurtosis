export function isDefined<T>(it: T | null | undefined): it is T {
  return it !== null && it !== undefined;
}

export function isStringTrue(value?: string | null) {
  return (value + "").toLowerCase() === "true";
}

export function assertDefined<T>(v: T | null | undefined, message: string = "Value is not defined"): asserts v is T {
  if (!isDefined(v)) {
    throw new Error(message);
  }
}

export function stringifyError(err: any): string {
  switch (typeof err) {
    case "bigint":
    case "boolean":
    case "number":
    case "string":
      return err.toString();

    case "object":
      if (err === null) {
        return "null";
      }
      if (err instanceof Error) {
        return err.toString();
      }
      return JSON.stringify(err);
    case "undefined":
      return "undefined";
    case "function":
      return "function";
    case "symbol":
      return "symbol";
  }
}

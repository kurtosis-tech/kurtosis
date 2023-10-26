import { Result } from "true-myth";

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

export type ErrorAndMessage<E> = {
  error: E;
  message?: string;
};

export async function asyncResult<T>(
  p: Promise<T> | (() => Promise<T>),
  errorMessage?: string,
): Promise<Result<T, ErrorAndMessage<any>>> {
  return (typeof p === "function" ? p() : p)
    .then((r) => Result.ok<T, ErrorAndMessage<any>>(r))
    .catch((err: any) => Result.err({ error: err, message: errorMessage }));
}

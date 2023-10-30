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

export function range(until: number): number[];
export function range(from: number, to: number): [];
export function range(from: number, to: number, step: number): number[];
export function range(a: number, b?: number, c?: number) {
  const start = isDefined(b) ? a : 0;
  const stop = isDefined(b) ? b : a;
  const step = isDefined(c) ? c : 1;
  return Array.from({ length: (stop - start) / step + 1 }, (_, i) => start + i * step);
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

export async function asyncResult<T>(
  p: Promise<T> | (() => Promise<T>),
  errorMessage?: string,
): Promise<Result<T, string>> {
  try {
    const r = await (typeof p === "function" ? p() : p);
    return Result.ok<T, string>(r);
  } catch (e: any) {
    return Result.err(errorMessage || stringifyError(e));
  }
}

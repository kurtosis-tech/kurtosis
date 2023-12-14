import { Result } from "true-myth";

export * from "./download";
export * from "./packageUtils";
export * from "./types";

export function isDefined<T>(it: T | null | undefined): it is T {
  return it !== null && it !== undefined;
}

export function isNotEmpty(it: string): it is string {
  return it.length > 0;
}

export function isStringTrue(value?: string | null) {
  return (value + "").toLowerCase() === "true";
}

export function assertDefined<T>(v: T | null | undefined, message: string = "Value is not defined"): asserts v is T {
  if (!isDefined(v)) {
    throw new Error(message);
  }
}

export function isIterable<T>(input: Iterable<T> | any): input is Iterable<T> {
  if (!isDefined(input)) {
    return false;
  }

  return typeof input[Symbol.iterator] === "function";
}

export function isAsyncIterable<T>(input: Iterable<T> | any): input is AsyncIterable<T> {
  if (!isDefined(input)) {
    return false;
  }

  return typeof input[Symbol.asyncIterator] === "function";
}

export function capitalize(input: string): string {
  return input
    .split(" ")
    .map((word) => (word.length >= 1 ? word[0].toUpperCase() + word.substring(1) : word))
    .join(" ");
}

const ansiPattern = [
  "[\\u001B\\u009B][[\\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]+)*|[a-zA-Z\\d]+(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\\u0007)",
  "(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-nq-uy=><~]))",
].join("|");

const ansiRegex = RegExp(ansiPattern, "g");

export function hasAnsi(text: string) {
  return ansiRegex.test(text);
}

export function stripAnsi(input: string): string {
  if (typeof input !== "string") {
    throw new TypeError(`Expected a \`string\`, got \`${typeof input}\``);
  }

  return input.replace(ansiRegex, "");
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

export function wrapResult<T>(c: () => T, errorMessage?: string): Result<T, string> {
  try {
    return Result.ok(c());
  } catch (e: any) {
    return Result.err(errorMessage || stringifyError(e));
  }
}

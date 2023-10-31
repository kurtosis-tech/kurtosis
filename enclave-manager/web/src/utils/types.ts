type NonFunctionKeyNames<T> = Exclude<
  {
    [key in keyof T]: T[key] extends Function ? never : key;
  }[keyof T],
  undefined
>;

export type RemoveFunctions<T> = Pick<T, NonFunctionKeyNames<T>>;

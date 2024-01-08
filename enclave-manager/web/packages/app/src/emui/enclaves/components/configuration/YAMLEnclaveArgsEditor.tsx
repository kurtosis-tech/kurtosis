import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { CodeEditor, isDefined, stringifyError } from "kurtosis-ui-components";
import { forwardRef, useImperativeHandle, useMemo, useState } from "react";
import YAML from "yaml";
import { transformFormArgsToKurtosisArgs } from "./utils";

type YAMLEditorProps = {
  kurtosisPackage: KurtosisPackage;
  onError: (error: string) => void;
  values?: Record<string, any>;
};
export type YAMLEditorImperativeAttributes = {
  getValues: () => Record<string, any> | null;
  isDirty: () => void;
};
export const YAMLEnclaveArgsEditor = forwardRef<YAMLEditorImperativeAttributes, YAMLEditorProps>(
  ({ kurtosisPackage, values, onError }: YAMLEditorProps, ref) => {
    const initText = useMemo(
      () => YAML.stringify(isDefined(values) ? transformFormArgsToKurtosisArgs(values, kurtosisPackage) : ""),
      [values, kurtosisPackage],
    );
    const [text, setText] = useState(initText);

    useImperativeHandle(
      ref,
      () => ({
        getValues: () => {
          try {
            const newValues = YAML.parse(text);
            if (typeof newValues !== "object") {
              throw new Error(`Expected text to be an object, but it's ${typeof newValues}`);
            }
            const invalidKeys = Object.keys(newValues).filter(
              (key) => !kurtosisPackage.args.some((arg) => arg.name === key),
            );
            if (invalidKeys.length > 0) {
              throw new Error(`Some of these keys are not valid for this package: ${invalidKeys.join(", ")}`);
            }
            // TODO: consider implementing yaml validation using the djv library
            return newValues;
          } catch (error: any) {
            onError(stringifyError(error));
          }
          return null;
        },
        isDirty: () => initText !== text,
      }),
      [initText, text, kurtosisPackage],
    );

    return <CodeEditor text={text} fileName={"config.yml"} onTextChange={setText} />;
  },
);

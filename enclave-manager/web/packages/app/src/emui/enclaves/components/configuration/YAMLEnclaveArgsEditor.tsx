import { Flex } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { CodeEditor, isDefined, KurtosisAlert, stringifyError } from "kurtosis-ui-components";
import { forwardRef, useImperativeHandle, useState } from "react";
import YAML from "yaml";
import { transformFormArgsToKurtosisArgs } from "./utils";

type YAMLEditorProps = {
  kurtosisPackage: KurtosisPackage;
  values?: Record<string, any>;
};
export type YAMLEditorImperativeAttributes = {
  getValues: () => Record<string, any> | null;
};
export const YAMLEnclaveArgsEditor = forwardRef<YAMLEditorImperativeAttributes, YAMLEditorProps>(
  ({ kurtosisPackage, values }: YAMLEditorProps, ref) => {
    const [text, setText] = useState(
      YAML.stringify(isDefined(values) ? transformFormArgsToKurtosisArgs(values, kurtosisPackage) : ""),
    );
    const [error, setError] = useState("");

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

            setError("");
            return newValues;
          } catch (error: any) {
            setError(stringifyError(error));
          }
          return null;
        },
      }),
      [text, kurtosisPackage],
    );

    return (
      <Flex flexDirection={"column"} gap={"10px"}>
        {error !== "" && <KurtosisAlert message={error} />}
        <CodeEditor text={text} fileName={"config.yml"} onTextChange={setText} />
      </Flex>
    );
  },
);

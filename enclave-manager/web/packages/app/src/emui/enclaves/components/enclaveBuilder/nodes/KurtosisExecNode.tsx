import { Flex } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo, useEffect } from "react";
import { useFormContext } from "react-hook-form";
import { NodeProps } from "reactflow";
import { IntegerArgumentInput } from "../../form/IntegerArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { KurtosisFormInputProps } from "../../form/types";
import { MentionStringArgumentInput } from "../input/MentionStringArgumentInput";
import { SelectServiceInput } from "../input/SelectServiceInput";
import { KurtosisExecNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";
import { KurtosisNode } from "./KurtosisNode";

export const KurtosisExecNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisExecNodeData;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNode id={id} selected={selected} minWidth={650} maxWidth={800}>
        <Flex gap={"16px"}>
          <ExecNameUpdater />
          <KurtosisFormControl<KurtosisExecNodeData>
            name={"service"}
            label={"Service"}
            isRequired
            isDisabled={nodeData.isFromPackage}
          >
            <SelectServiceInput isRequired disabled={nodeData.isFromPackage} name={"service"} />
          </KurtosisFormControl>
        </Flex>
        <Flex flexDirection={"column"} gap={"8px"}>
          <KurtosisFormControl<KurtosisExecNodeData>
            name={"command"}
            label={"Command"}
            isRequired
            isDisabled={nodeData.isFromPackage}
          >
            <MentionStringArgumentInput
              size={"sm"}
              name={"command"}
              multiline
              isRequired
              disabled={nodeData.isFromPackage}
            />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisExecNodeData>
            name={"acceptableCodes"}
            label={"Acceptable Exit Codes"}
            isDisabled={nodeData.isFromPackage}
            helperText={"If the executed command returns a code not on this list starlark will fail. Defaults to [0]"}
          >
            <ListArgumentInput<KurtosisExecNodeData>
              FieldComponent={AcceptableCodeInput}
              size={"sm"}
              name={"acceptableCodes"}
              createNewValue={() => ({ value: 0 })}
              disabled={nodeData.isFromPackage}
            />
          </KurtosisFormControl>
        </Flex>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);

const AcceptableCodeInput = (props: KurtosisFormInputProps<KurtosisExecNodeData>) => {
  return (
    <IntegerArgumentInput<KurtosisExecNodeData>
      {...props}
      size={"sm"}
      name={`${props.name as `acceptableCodes.${number}`}.value`}
    />
  );
};

const ExecNameUpdater = memo(() => {
  const { variables } = useVariableContext();
  const { watch, setValue } = useFormContext();

  const service = watch("service");
  const name = watch("name");

  useEffect(() => {
    const serviceVariableId = service.replace(/\{\{(.*)}}/, "$1");
    const serviceName = variables.find((v) => v.id === serviceVariableId)?.displayName || "Unknown";
    if (name !== `${serviceName} exec`) {
      setValue("name", `${serviceName} exec`);
    }
  }, [name, service, setValue, variables]);

  return null;
});

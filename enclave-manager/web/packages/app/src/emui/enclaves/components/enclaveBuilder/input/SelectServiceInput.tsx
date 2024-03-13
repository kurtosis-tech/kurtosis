import { useMemo } from "react";
import { SelectArgumentInput, SelectOption } from "../../form/SelectArgumentInput";
import { KurtosisFormInputProps } from "../../form/types";
import { KurtosisExecNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";

export const SelectServiceInput = (props: KurtosisFormInputProps<KurtosisExecNodeData>) => {
  const { variables } = useVariableContext();
  const serviceVariableOptions = useMemo((): SelectOption[] => {
    return variables
      .filter((variable) => variable.id.match(/^(?:service)\.[^.]+\.name$/))
      .map((variable) => ({ display: variable.displayName, value: `{{${variable.id}}}` }));
  }, [variables]);

  return (
    <SelectArgumentInput<KurtosisExecNodeData>
      options={serviceVariableOptions}
      {...props}
      size={"sm"}
      placeholder={"Select a Service"}
      name={props.name}
    />
  );
};

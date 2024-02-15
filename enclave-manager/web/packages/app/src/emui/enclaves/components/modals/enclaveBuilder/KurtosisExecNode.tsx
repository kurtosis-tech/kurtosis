import { Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo, useMemo } from "react";
import { NodeProps } from "reactflow";
import { IntegerArgumentInput } from "../../form/IntegerArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { SelectArgumentInput, SelectOption } from "../../form/SelectArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../../form/types";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import { validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisExecNodeData, KurtosisServiceNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

export const KurtosisExecNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data, variables } = useVariableContext();
    const nodeData = data[id] as KurtosisExecNodeData;

    const serviceVariableOptions = useMemo((): SelectOption[] => {
      return variables
        .filter((variable) => variable.id.match(/^service\.[^.]+\.name+$/))
        .map((variable) => ({
          display: variable.displayName.replace(/service\.(.*)\.name/, "$1"),
          value: `{{${variable.id}}}`,
        }));
    }, [variables]);

    if (!isDefined(nodeData)) {
      // Node has probably been deleted.
      return null;
    }

    return (
      <KurtosisNode
        id={id}
        selected={selected}
        name={nodeData.execName}
        color={"purple.900"}
        minWidth={300}
        maxWidth={800}
      >
        <KurtosisFormControl<KurtosisExecNodeData> name={"execName"} label={"Exec Name"} isRequired>
          <StringArgumentInput size={"sm"} name={"execName"} isRequired validate={validateName} />
        </KurtosisFormControl>
        <Tabs>
          <TabList>
            <Tab>Config</Tab>
            <Tab>Advanced</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              {" "}
              <KurtosisFormControl<KurtosisServiceNodeData>
                name={"serviceName"}
                label={"Service"}
                helperText={"Choose which service to run this command in."}
                isRequired
              >
                <SelectArgumentInput<KurtosisServiceNodeData>
                  options={serviceVariableOptions}
                  isRequired
                  size={"sm"}
                  placeholder={"Select a Service"}
                  name={`serviceName`}
                />
              </KurtosisFormControl>
              <KurtosisFormControl<KurtosisExecNodeData> name={"command"} label={"Command"} isRequired>
                <MentionStringArgumentInput size={"sm"} name={"command"} isRequired />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisExecNodeData>
                name={"acceptableCodes"}
                label={"Acceptable Exit Codes"}
                isRequired
              >
                <ListArgumentInput<KurtosisExecNodeData>
                  FieldComponent={AcceptableCodeInput}
                  size={"sm"}
                  name={"acceptableCodes"}
                  createNewValue={() => ({ value: 0 })}
                  isRequired
                />
              </KurtosisFormControl>
            </TabPanel>
          </TabPanels>
        </Tabs>
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

import { isDefined } from "kurtosis-ui-components";
import { useCallback, useMemo } from "react";
import { Controller } from "react-hook-form";
import { Mention, MentionsInput } from "react-mentions";
import { useNodeId } from "reactflow";
import { KurtosisFormInputProps } from "../../form/types";
import { useVariableContext } from "../VariableContextProvider";
import "./MentionStringArgumentInput.css";

type MentionStringArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel> & {
  multiline?: boolean;
};

export const MentionStringArgumentInput = <DataModel extends object>({
  name,
  placeholder,
  isRequired,
  validate,
  disabled,
  width,
  tabIndex,
  multiline,
}: MentionStringArgumentInputProps<DataModel>) => {
  const { variables, data } = useVariableContext();
  const nodeId = useNodeId();

  const allowedVariables = useMemo(() => {
    if (!isDefined(nodeId)) {
      return variables;
    }
    const nodeData = data[nodeId];
    // Exclude the variable currently being edited
    const excludedVariable = `${nodeData.type}.${nodeId}.${name}`;
    return variables.filter((v) => v.id !== excludedVariable);
  }, [data, nodeId, name, variables]);

  const handleQuery = useCallback(
    (query?: string) => {
      if (!isDefined(query)) {
        return [];
      }
      const suggestions = allowedVariables.map((v) => ({ display: v.displayName, id: v.id }));
      const queryTerms = query.toLowerCase().split(/\s+|\./);
      return suggestions.filter((variable) =>
        queryTerms.every((term) => variable.display.toLowerCase().includes(term)),
      );
    },
    [allowedVariables],
  );

  return (
    <Controller
      name={name}
      defaultValue={"" as any}
      rules={{ required: isRequired, validate: validate }}
      render={({ field, fieldState }) => {
        return (
          <MentionsInput
            placeholder={placeholder}
            className={"mentions"}
            style={{
              "&singleLine": {
                width: width,
              },
              "&multiLine": {
                minHeight: "90px",
                overflow: "scroll",
              },
              maxWidth: "600px",
            }}
            aria-invalid={fieldState.invalid}
            tabIndex={tabIndex}
            singleLine={!multiline}
            value={field.value}
            disabled={disabled}
            onChange={(e, newValue, newPlainTextValue, mentions) => field.onChange(newValue)}
          >
            <Mention
              className={"mentions__mention"}
              trigger={/(?<=^|.*[ :/@#$])(([^ :/@#$]{2,}))$/}
              markup={"{{__id__}}"}
              data={handleQuery}
              displayTransform={(id) =>
                variables.find((variable) => variable.id === id)?.displayName || "Missing Variable"
              }
            />
          </MentionsInput>
        );
      }}
    />
  );
};

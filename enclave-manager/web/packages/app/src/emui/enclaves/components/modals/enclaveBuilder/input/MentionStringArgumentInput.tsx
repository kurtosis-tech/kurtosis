import { isDefined } from "kurtosis-ui-components";
import { useCallback } from "react";
import { Controller } from "react-hook-form";
import { Mention, MentionsInput } from "react-mentions";
import { KurtosisFormInputProps } from "../../../form/types";
import { useVariableContext } from "../VariableContextProvider";
import "./MentionStringArgumentInput.css";

type MentionStringArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel>;

export const MentionStringArgumentInput = <DataModel extends object>({
  name,
  placeholder,
  isRequired,
  validate,
  disabled,
  width,
  size,
  tabIndex,
}: MentionStringArgumentInputProps<DataModel>) => {
  const { variables } = useVariableContext();

  const handleQuery = useCallback(
    (query?: string) => {
      if (!isDefined(query)) {
        return [];
      }
      const suggestions = variables.map((v) => ({ display: v.displayName, id: v.id }));
      const queryTerms = query.toLowerCase().split(/\s+|\./);
      return suggestions.filter((variable) => queryTerms.every((term) => variable.display.includes(term)));
    },
    [variables],
  );

  return (
    <Controller
      name={name}
      disabled={disabled}
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
            }}
            aria-invalid={fieldState.invalid}
            tabIndex={tabIndex}
            singleLine
            value={field.value}
            onChange={(e, newValue, newPlainTextValue, mentions) => field.onChange(newValue)}
          >
            <Mention
              className={"mentions__mention"}
              trigger={/((?:@)?(\S\S.*))$/}
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

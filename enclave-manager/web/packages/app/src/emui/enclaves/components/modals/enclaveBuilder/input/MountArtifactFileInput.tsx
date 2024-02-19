import { Grid, GridItem } from "@chakra-ui/react";
import { useMemo } from "react";
import { SelectArgumentInput, SelectOption } from "../../../form/SelectArgumentInput";
import { StringArgumentInput } from "../../../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../../../form/types";
import { KurtosisServiceNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";

export const MountArtifactFileInput = (props: KurtosisFormInputProps<KurtosisServiceNodeData>) => {
  const { variables } = useVariableContext();
  const artifactVariableOptions = useMemo((): SelectOption[] => {
    return variables
      .filter((variable) => variable.id.match(/^(?:artifact|shell|python)\.[^.]+$/))
      .map((variable) => ({ display: variable.displayName, value: `{{${variable.id}}}` }));
  }, [variables]);

  return (
    <Grid gridTemplateColumns={"1fr 1fr"} gridGap={"8px"} p={"8px"} bgColor={"gray.650"}>
      <GridItem>
        <StringArgumentInput<KurtosisServiceNodeData>
          {...props}
          size={"sm"}
          placeholder={"/some/path"}
          name={`${props.name as `files.${number}`}.mountPoint`}
        />
      </GridItem>
      <GridItem>
        <SelectArgumentInput<KurtosisServiceNodeData>
          options={artifactVariableOptions}
          {...props}
          size={"sm"}
          placeholder={"Select an Artifact"}
          name={`${props.name as `files.${number}`}.artifactName`}
        />
      </GridItem>
    </Grid>
  );
};

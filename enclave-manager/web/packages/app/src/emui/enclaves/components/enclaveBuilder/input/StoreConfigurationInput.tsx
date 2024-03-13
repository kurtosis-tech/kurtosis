import { Grid, GridItem } from "@chakra-ui/react";
import { Path } from "react-hook-form";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../../form/types";
import { KurtosisStore } from "../types";
import { MentionStringArgumentInput } from "./MentionStringArgumentInput";

export const StoreConfigurationInput = <F extends { store: KurtosisStore[] }>(props: KurtosisFormInputProps<F>) => {
  return (
    <Grid gridTemplateColumns={"1fr 1fr"} gridGap={"8px"} p={"8px"} bgColor={"gray.650"}>
      <GridItem>
        <StringArgumentInput
          {...props}
          size={"sm"}
          placeholder={"store name"}
          name={`${props.name as `store.${number}`}.name` as Path<F>}
        />
      </GridItem>
      <GridItem>
        <MentionStringArgumentInput<F>
          {...props}
          size={"sm"}
          placeholder={"/some/path"}
          name={`${props.name as `store.${number}`}.path` as Path<F>}
        />
      </GridItem>
    </Grid>
  );
};

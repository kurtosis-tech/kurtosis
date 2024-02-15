import { Grid, GridItem } from "@chakra-ui/react";
import { IntegerArgumentInput } from "../../../form/IntegerArgumentInput";
import { OptionsArgumentInput } from "../../../form/OptionArgumentInput";
import { StringArgumentInput } from "../../../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../../../form/types";
import { KurtosisServiceNodeData } from "../types";

export const PortConfigurationField = (props: KurtosisFormInputProps<KurtosisServiceNodeData>) => (
  <Grid gridTemplateColumns={"1fr 1fr"} gridGap={"8px"} p={"8px"} bgColor={"gray.650"}>
    <GridItem>
      <StringArgumentInput<KurtosisServiceNodeData>
        {...props}
        size={"sm"}
        placeholder={"Port Name (eg postgres)"}
        name={`${props.name as `ports.${number}`}.portName`}
      />
    </GridItem>
    <GridItem>
      <StringArgumentInput<KurtosisServiceNodeData>
        {...props}
        size={"sm"}
        placeholder={"Application Protocol (eg postgresql)"}
        name={`${props.name as `ports.${number}`}.applicationProtocol`}
        validate={(val) => {
          if (typeof val !== "string") {
            return "Value should be a string";
          }
          if (val.includes(" ")) {
            return "Application protocol cannot include spaces";
          }
        }}
      />
    </GridItem>
    <GridItem>
      <OptionsArgumentInput<KurtosisServiceNodeData>
        {...props}
        options={["TCP", "UDP"]}
        name={`${props.name as `ports.${number}`}.transportProtocol`}
      />
    </GridItem>
    <GridItem>
      <IntegerArgumentInput<KurtosisServiceNodeData>
        {...props}
        name={`${props.name as `ports.${number}`}.port`}
        size={"sm"}
      />
    </GridItem>
  </Grid>
);

import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { FileTreeArgumentInput } from "./input/FileTreeArgumentInput";
import { validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisArtifactNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

export const KurtosisArtifactNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisArtifactNodeData;

    if (!isDefined(nodeData)) {
      // Node has probably been deleted.
      return null;
    }

    return (
      <KurtosisNode
        id={id}
        selected={selected}
        name={nodeData.artifactName}
        color={"yellow.900"}
        minWidth={300}
        maxWidth={800}
      >
        <KurtosisFormControl<KurtosisArtifactNodeData> name={"artifactName"} label={"Artifact Name"} isRequired>
          <StringArgumentInput size={"sm"} name={"artifactName"} isRequired validate={validateName} />
        </KurtosisFormControl>
        <KurtosisFormControl name={"files"} label={"Files"}>
          <FileTreeArgumentInput name={"files"} />
        </KurtosisFormControl>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id !== newProps.id || oldProps.selected !== newProps.selected,
);

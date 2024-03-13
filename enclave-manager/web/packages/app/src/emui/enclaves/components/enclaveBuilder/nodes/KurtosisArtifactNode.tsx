import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { FileTreeArgumentInput } from "../input/FileTreeArgumentInput";
import { validateName } from "../input/validators";
import { KurtosisArtifactNodeData, KurtosisPythonNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";
import { KurtosisNode } from "./KurtosisNode";

export const KurtosisArtifactNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisPythonNodeData;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNode id={id} selected={selected} minWidth={300} maxWidth={800}>
        <KurtosisFormControl<KurtosisArtifactNodeData>
          name={"name"}
          label={"Artifact Name"}
          isRequired
          isDisabled={nodeData.isFromPackage}
        >
          <StringArgumentInput
            size={"sm"}
            name={"name"}
            isRequired
            validate={validateName}
            disabled={nodeData.isFromPackage}
          />
        </KurtosisFormControl>
        <KurtosisFormControl name={"files"} label={"Files"} isDisabled={nodeData.isFromPackage}>
          <FileTreeArgumentInput name={"files"} disabled={nodeData.isFromPackage} />
        </KurtosisFormControl>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);

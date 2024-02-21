import { memo } from "react";
import { NodeProps } from "reactflow";
import { KurtosisFormControl } from "../form/KurtosisFormControl";
import { StringArgumentInput } from "../form/StringArgumentInput";
import { FileTreeArgumentInput } from "./input/FileTreeArgumentInput";
import { validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisArtifactNodeData } from "./types";

export const KurtosisArtifactNode = memo(
  ({ id, selected }: NodeProps) => {
    return (
      <KurtosisNode id={id} selected={selected} minWidth={300} maxWidth={800}>
        <KurtosisFormControl<KurtosisArtifactNodeData> name={"artifactName"} label={"Artifact Name"} isRequired>
          <StringArgumentInput size={"sm"} name={"artifactName"} isRequired validate={validateName} />
        </KurtosisFormControl>
        <KurtosisFormControl name={"files"} label={"Files"}>
          <FileTreeArgumentInput name={"files"} />
        </KurtosisFormControl>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);

import { Button, Flex, Text } from "@chakra-ui/react";
import React, { useCallback, useMemo, useState } from "react";
import { AiFillFile, AiFillFolder, AiFillFolderOpen } from "react-icons/ai";
import { FileSize } from "./FileSize";
import { isDefined } from "./utils";

/**
 * This file tree component recursively renders itself to present a file tree.
 * To keep this performant the nodes (DirectoryNode and FileNode) must make use of
 * useCallback and useMemo. This allows the React.memo around FileTreeMode to function
 * and skip rendering unchanged components.
 */

export type FileTreeNode = {
  name: string;
  size?: bigint;
  childNodes?: FileTreeNode[];
};

type FileTreeProps = {
  nodes: FileTreeNode[];
  selectedFilePath?: string[];
  onFileSelected: (selectedFilePath: string[]) => void;
  // Internal prop used for padding
  _isChildNode?: boolean;
};

export const FileTree = ({ nodes, selectedFilePath, onFileSelected, _isChildNode }: FileTreeProps) => {
  return (
    <Flex flexDirection={"column"} pl={_isChildNode ? "22px" : undefined} w={"100%"}>
      {nodes.map((node, i) => (
        <FileTreeNodeComponent
          key={node.name}
          node={node}
          selectedFilePath={
            isDefined(selectedFilePath) && selectedFilePath.length > 0 && selectedFilePath[0] === node.name
              ? selectedFilePath
              : undefined
          }
          onFileSelected={onFileSelected}
        />
      ))}
    </Flex>
  );
};

type FileTreeNodeComponentProps = {
  node: FileTreeNode;
  selectedFilePath?: string[];
  onFileSelected: (selectedFilePath: string[]) => void;
};

const FileTreeNodeComponent = React.memo((props: FileTreeNodeComponentProps) => {
  if (isDefined(props.node.childNodes)) {
    return <DirectoryNode {...(props as FileTreeNodeComponentProps & { node: { childNodes: FileTreeNode[] } })} />;
  } else {
    return <FileNode {...props} />;
  }
});

const DirectoryNode = ({
  node,
  selectedFilePath,
  onFileSelected,
}: FileTreeNodeComponentProps & { node: { childNodes: FileTreeNode[] } }) => {
  const [collapsed, setCollapsed] = useState(false);

  const childSelectedFilePath = useMemo(
    () =>
      isDefined(selectedFilePath) && selectedFilePath.length > 0 && selectedFilePath[0] === node.name
        ? selectedFilePath.slice(1)
        : undefined,
    [selectedFilePath, node],
  );

  const handleClick = useCallback(() => {
    setCollapsed((collapsed) => !collapsed);
  }, []);

  const handleFileSelected = useCallback(
    (filePath: string[]) => onFileSelected([node.name, ...filePath]),
    [onFileSelected, node],
  );

  return (
    <>
      <Button
        variant={"fileTree"}
        size={"xs"}
        onClick={handleClick}
        leftIcon={collapsed ? <AiFillFolder /> : <AiFillFolderOpen />}
      >
        {node.name}
      </Button>
      {!collapsed && (
        <FileTree
          nodes={node.childNodes}
          onFileSelected={handleFileSelected}
          selectedFilePath={childSelectedFilePath}
          _isChildNode
        />
      )}
    </>
  );
};

const FileNode = ({ node, selectedFilePath, onFileSelected }: FileTreeNodeComponentProps) => {
  const isSelected = isDefined(selectedFilePath) && selectedFilePath.length === 1 && selectedFilePath[0] === node.name;
  return (
    <Button
      variant={"fileTree"}
      size={"xs"}
      leftIcon={<AiFillFile />}
      rightIcon={<FileSize fileSize={node.size} color={"gray.250"} />}
      isActive={isSelected}
      onClick={() => onFileSelected([node.name])}
    >
      <Text as={"span"} w={"100%"} textAlign={"left"}>
        {node.name}
      </Text>
    </Button>
  );
};

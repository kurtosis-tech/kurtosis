import { Box, Button, ButtonGroup, Flex, Icon, Menu, MenuButton, MenuItem, MenuList } from "@chakra-ui/react";
import { useRef } from "react";
import { FiDownload, FiGrid, FiPlus } from "react-icons/fi";
import { Node, useOnViewportChange, useReactFlow, XYPosition } from "reactflow";
import { v4 as uuidv4 } from "uuid";
import { nodeIcons } from "./nodes/KurtosisNode";
import { useUIState } from "./UIStateContext";
import { useVariableContext } from "./VariableContextProvider";

export const Toolbar = () => {
  const insertOffset = useRef(0);
  const { updateData } = useVariableContext();
  const { getViewport, addNodes } = useReactFlow();

  const { applyAutoLayout, toggleExpanded, zoomToNode } = useUIState();

  useOnViewportChange({ onEnd: () => (insertOffset.current = 1) });

  const getNewNodePosition = (): XYPosition => {
    const viewport = getViewport();
    insertOffset.current += 1;
    return { x: -viewport.x + insertOffset.current * 20 + 400, y: -viewport.y + insertOffset.current * 20 };
  };

  const addAndFocusNode = (node: Node) => {
    addNodes(node);
    toggleExpanded(node.id);
    setTimeout(() => zoomToNode(node), 0);
  };

  const handleAddServiceNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "service",
      name: "",
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
      ports: [],
      env: [],
      files: [],
      cmd: "",
      entrypoint: "",
      isValid: false,
    });
    addAndFocusNode({
      id,
      position: getNewNodePosition(),
      width: 650,
      style: { width: "650px" },
      type: "serviceNode",
      data: {},
    });
  };

  const handleAddExecNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "exec",
      name: "",
      service: "",
      command: "",
      acceptableCodes: [],
      isValid: false,
    });
    addAndFocusNode({
      id,
      position: getNewNodePosition(),
      width: 650,
      style: { width: "650px" },
      type: "execNode",
      data: {},
    });
  };

  const handleAddArtifactNode = () => {
    const id = uuidv4();
    updateData(id, { type: "artifact", name: "", files: {}, isValid: false });
    addAndFocusNode({
      id,
      position: getNewNodePosition(),
      width: 400,
      style: { width: "400px" },
      type: "artifactNode",
      data: {},
    });
  };

  const handleAddShellNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "shell",
      name: "",
      command: "",
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
      env: [],
      files: [],
      store: [{ name: "", path: "" }],
      wait_enabled: "true",
      wait: "",
      isValid: false,
    });
    addAndFocusNode({
      id,
      position: getNewNodePosition(),
      width: 650,
      style: { width: "650px" },
      type: "shellNode",
      data: {},
    });
  };

  const handleAddPythonNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "python",
      name: "",
      command: "",
      packages: [],
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
      args: [],
      files: [],
      store: [{ name: "", path: "" }],
      wait_enabled: "true",
      wait: "",
      isValid: false,
    });
    addAndFocusNode({
      id,
      position: getNewNodePosition(),
      width: 650,
      style: { width: "650px" },
      type: "pythonNode",
      data: {},
    });
  };

  const handleAddPackageNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "package",
      name: "",
      packageId: "",
      args: {},
      locator: "",
      isValid: false,
    });
    addAndFocusNode({
      id,
      selected: true,
      position: getNewNodePosition(),
      width: 900,
      style: { width: "900px" },
      type: "packageNode",
      data: {},
    });
  };

  return (
    <Box
      borderRadius={"5px"}
      position={"absolute"}
      zIndex={"99999"}
      top={"20px"}
      left={"20px"}
      bg={"gray.800"}
      p={"8px"}
    >
      <ButtonGroup size={"sm"}>
        <Menu>
          <MenuButton as={Button} leftIcon={<FiPlus />}>
            Add node
          </MenuButton>
          <MenuList>
            <MenuItem as="button" onClick={handleAddServiceNode}>
              <Flex gap={2} alignItems="center">
                <Icon as={nodeIcons["service"]} /> Service
              </Flex>
            </MenuItem>
            <MenuItem as="button" onClick={handleAddArtifactNode}>
              <Flex gap={2} alignItems="center">
                <Icon as={nodeIcons["artifact"]} /> File
              </Flex>
            </MenuItem>
            <MenuItem as="button" onClick={handleAddExecNode}>
              <Flex gap={2} alignItems="center">
                <Icon as={nodeIcons["exec"]} /> Exec Task
              </Flex>
            </MenuItem>
            <MenuItem as="button" onClick={handleAddShellNode}>
              <Flex gap={2} alignItems="center">
                <Icon as={nodeIcons["shell"]} /> Shell Script
              </Flex>
            </MenuItem>
            <MenuItem as="button" onClick={handleAddPythonNode}>
              <Flex gap={2} alignItems="center">
                <Icon as={nodeIcons["python"]} /> Python Script
              </Flex>
            </MenuItem>
          </MenuList>
        </Menu>
        <Button leftIcon={<FiDownload />} onClick={handleAddPackageNode}>
          Import Package
        </Button>
        <Button leftIcon={<FiGrid />} onClick={applyAutoLayout}>
          Apply Auto-Layout
        </Button>
      </ButtonGroup>
    </Box>
  );
};

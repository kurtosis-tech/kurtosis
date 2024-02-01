import {
  Button,
  Editable,
  EditableInput,
  EditablePreview,
  Flex,
  FormControl,
  FormLabel, Grid, GridItem,
  IconButton,
  Input,
} from "@chakra-ui/react";
import {Fragment, memo} from "react";
import {FiDelete, FiPlus, FiTrash} from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { NodeProps, NodeResizeControl, useReactFlow } from "reactflow";

type KurtosisPort = {
  port: number;
  transportProtocol: "TCP" | "UDP";
  applicationProtocol: string;
};

export type KurtosisServiceNodeData = {
  name: string;
  image: string;
  env: { key: string; value: string }[];
  ports: KurtosisPort[];
};

export const KurtosisServiceNode = memo(({ id, data, isConnectable, selected }: NodeProps<KurtosisServiceNodeData>) => {
  const { deleteElements, setNodes } = useReactFlow<KurtosisServiceNodeData>();

  const handleDeleteNode = () => {
    deleteElements({ nodes: [{ id }] });
  };

  const handleDataUpdate = <K extends keyof KurtosisServiceNodeData>(key: K, value: KurtosisServiceNodeData[K]) => {
    setNodes((nodes) =>
      nodes.map((node) => (node.id !== id ? node : { ...node, data: { ...node.data, [key]: value } })),
    );
  };

  return (
    <Flex
      flexDirection={"column"}
      height={"100%"}
      borderRadius={"8px"}
      p={selected ? "8px" : "10px"}
      bg={"gray.600"}
      borderWidth={selected ? "3px" : "1px"}
      borderColor={selected ? "gray.850" : "gray.850"}
    >
      {/*<Handle*/}
      {/*  type="target"*/}
      {/*  position={Position.Left}*/}
      {/*  style={{ background: "#555" }}*/}
      {/*  onConnect={(params) => console.log("handle onConnect", params)}*/}
      {/*  isConnectable={isConnectable}*/}
      {/*/>*/}
      <NodeResizeControl minWidth={200} minHeight={100} style={{ background: "transparent", border: "none" }}>
        <RxCornerBottomRight style={{ position: "absolute", right: 5, bottom: 5 }} />
      </NodeResizeControl>

      <Flex justifyContent={"space-between"}>
        <Editable
          fontSize={"md"}
          fontWeight={"semibold"}
          value={data.name}
          onChange={(e) => handleDataUpdate("name", e)}
        >
          <EditablePreview />
          <EditableInput />
        </Editable>
        <IconButton
          aria-label={"Delete node"}
          icon={<FiTrash />}
          colorScheme={"red"}
          variant={"ghost"}
          size={"xs"}
          onClick={handleDeleteNode}
        />
      </Flex>
      <Flex flexDirection={"column"} bg={"gray.800"} p={"5px 16px"} flex={"1"}>
        <FormControl>
          <FormLabel fontSize={"xs"} fontWeight={"bold"}>
            Container
          </FormLabel>
          <Input size={"xs"} value={data.image} onChange={(e) => handleDataUpdate("image", e.target.value)} />
        </FormControl>
        <FormControl>
          <FormLabel fontSize={"xs"} fontWeight={"bold"}>
            Environment Variables
          </FormLabel>
          <Grid gridTemplateColumns={"1fr 1fr auto"} gridGap={"6px"}>
          {data.env.map(({key, value}, i) => <Fragment key={i}>
            <GridItem>
              <Input value={key} size={"sm"} onChange={e => handleDataUpdate('env', data.env.map((envVar, j) => i !== j ? envVar : {...envVar, key: e.target.value}))}/>
            </GridItem> <GridItem>
            <Input value={value} size={"sm"} onChange={e => handleDataUpdate('env', data.env.map((envVar, j) => i !== j ? envVar : {...envVar, value: e.target.value}))}/>
          </GridItem>
            <Button onClick={() => handleDataUpdate('env', data.env.filter((_, j) => i !== j))} leftIcon={<FiDelete />} size={"sm"} colorScheme={"red"} variant={"outline"}>
              Delete
            </Button>
          </Fragment>)}
          </Grid>
          <Button
              onClick={() => handleDataUpdate("env", [...data.env, {key: "", value: ""}])}
              leftIcon={<FiPlus />}
              size={"sm"}
              colorScheme={"kurtosisGreen"}
              variant={"outline"}
          >
            Add
          </Button>
        </FormControl>
      </Flex>
    </Flex>
  );
});

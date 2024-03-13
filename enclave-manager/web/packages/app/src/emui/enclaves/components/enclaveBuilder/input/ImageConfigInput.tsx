import {
  Code,
  Flex,
  IconButton,
  InputGroup,
  InputLeftAddon,
  InputRightElement,
  Popover,
  PopoverArrow,
  PopoverBody,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
  Portal,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
} from "@chakra-ui/react";

import { useState } from "react";
import { useFormContext } from "react-hook-form";
import { FiSettings } from "react-icons/fi";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { KurtosisImageConfig, KurtosisImageType } from "../types";
import { validateDockerLocator } from "./validators";

const tabs: { display: string; value: KurtosisImageType }[] = [
  { display: "Image", value: "image" },
  { display: "Dockerfile", value: "dockerfile" },
  { display: "Nix", value: "nix" },
];

type ImageConfigInputProps = {
  disabled?: boolean;
};
export const ImageConfigInput = ({ disabled }: ImageConfigInputProps) => {
  const { setValue, watch } = useFormContext<{ image: KurtosisImageConfig }>();
  const imageName = watch("image.image");
  const imageType = watch("image.type");
  const [activeTabIndex, setActiveTabIndex] = useState(tabs.findIndex((v) => v.value === imageType));

  const handleTabsChange = (newTabIndex: number) => {
    setActiveTabIndex(newTabIndex);
    if (!disabled) {
      setValue("image.type", tabs[activeTabIndex].value);
    }
  };

  return (
    <InputGroup size={"sm"}>
      <InputLeftAddon>{tabs.find((t) => t.value === imageType)?.display || "Unknown"}</InputLeftAddon>
      <StringArgumentInput
        name={"image.image"}
        validate={validateDockerLocator}
        placeholder={"Default: python:3.11-alpine"}
        paddingInlineEnd={8}
        disabled={disabled}
      />
      <InputRightElement>
        <Popover placement={"right-end"} isLazy>
          <PopoverTrigger>
            <IconButton
              icon={<FiSettings />}
              bg={"gray.850"}
              aria-label={"Image Configuration"}
              variant={"ghost"}
              size={"xs"}
            />
          </PopoverTrigger>
          <Portal>
            <PopoverContent fontSize={"sm"} minW={"500px"} minH={"520px"}>
              <PopoverArrow />
              <PopoverHeader>Image Configuration</PopoverHeader>
              <PopoverBody>
                <Flex flexDirection={"column"} gap={"8px"}>
                  <Text>
                    Configuration for the container with <Code>{imageName}</Code>
                  </Text>
                  <Text>Select the image type:</Text>
                  <Tabs index={activeTabIndex} onChange={handleTabsChange}>
                    <TabList>
                      {tabs.map((tab, i) => (
                        <Tab key={i}>{tab.display}</Tab>
                      ))}
                    </TabList>
                    <TabPanels>
                      <TabPanel>
                        <KurtosisFormControl
                          size={"xs"}
                          name={"image.username"}
                          label={"Username"}
                          isDisabled={disabled}
                          helperText={"The username that will be used to pull the image from the given registry"}
                        >
                          <StringArgumentInput size={"xs"} name={"imageConfig.username"} disabled={disabled} />
                        </KurtosisFormControl>
                        <KurtosisFormControl
                          name={"image.password"}
                          label={"Username"}
                          size={"xs"}
                          isDisabled={disabled}
                          helperText={"The pasword that will be used to pull the image from the given registry"}
                        >
                          <StringArgumentInput
                            name={"imageConfig.password"}
                            size={"xs"}
                            type={"password"}
                            disabled={disabled}
                          />
                        </KurtosisFormControl>
                        <KurtosisFormControl
                          size={"xs"}
                          name={"image.registry"}
                          label={"Registry"}
                          isDisabled={disabled}
                          helperText={"The URL of the registry"}
                        >
                          <StringArgumentInput
                            name={"image.username"}
                            size={"xs"}
                            placeholder={"http://my.registry.io"}
                            disabled={disabled}
                          />
                        </KurtosisFormControl>
                      </TabPanel>
                      <TabPanel>
                        <KurtosisFormControl
                          size={"xs"}
                          name={"image.buildContextDir"}
                          label={"Build Context Dir"}
                          helperText={
                            "Locator to build context within the Kurtosis package. As of now, Kurtosis expects a Dockerfile at the root of the build context"
                          }
                          isDisabled={disabled}
                          isRequired={activeTabIndex === 1}
                        >
                          <StringArgumentInput
                            size={"xs"}
                            name={"imageConfig.buildContextDir"}
                            isRequired={activeTabIndex === 1}
                            disabled={disabled}
                          />
                        </KurtosisFormControl>
                        <KurtosisFormControl
                          name={"imageConfig.targetStage"}
                          label={"Target Stage"}
                          size={"xs"}
                          isDisabled={disabled}
                          helperText={"Stage of image build to target for multi-stage container image"}
                        >
                          <StringArgumentInput name={"imageConfig.targetStage"} size={"xs"} disabled={disabled} />
                        </KurtosisFormControl>
                      </TabPanel>
                      <TabPanel>
                        <KurtosisFormControl
                          size={"xs"}
                          name={"image.buildContextDir"}
                          label={"Build Context Dir"}
                          helperText={"Locator to build context within the Kurtosis package."}
                          isDisabled={disabled}
                          isRequired={activeTabIndex === 2}
                        >
                          <StringArgumentInput
                            size={"xs"}
                            name={"image.buildContextDir"}
                            isRequired={activeTabIndex === 2}
                            disabled={disabled}
                            placeholder={"./"}
                          />
                        </KurtosisFormControl>
                        <KurtosisFormControl
                          name={"image.flakeLocationDir"}
                          label={"Flake Location Dir"}
                          size={"xs"}
                          helperText={
                            "The relative path (from the `build_context_dir`) to the folder containing the flake.nix file"
                          }
                          isDisabled={disabled}
                          isRequired={activeTabIndex === 2}
                        >
                          <StringArgumentInput
                            name={"image.flakeLocationDir"}
                            size={"xs"}
                            placeholder={"./hello-go"}
                            disabled={disabled}
                            isRequired={activeTabIndex === 2}
                          />
                        </KurtosisFormControl>
                        <KurtosisFormControl
                          name={"image.flakeOutput"}
                          label={"Flake Output"}
                          size={"xs"}
                          isDisabled={disabled}
                          helperText={
                            "The selector for the Flake output with the image derivation. Fallbacks to the default package."
                          }
                        >
                          <StringArgumentInput name={"image.flakeOutput"} size={"xs"} disabled={disabled} />
                        </KurtosisFormControl>
                      </TabPanel>
                    </TabPanels>
                  </Tabs>
                </Flex>
              </PopoverBody>
            </PopoverContent>
          </Portal>
        </Popover>
      </InputRightElement>
    </InputGroup>
  );
};

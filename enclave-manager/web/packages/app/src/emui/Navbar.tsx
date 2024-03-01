import {
  Button,
  Input,
  InputGroup,
  InputRightElement,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
} from "@chakra-ui/react";
import { CopyButton, NavButton, Navigation, NavigationDivider } from "kurtosis-ui-components";
import { useState } from "react";
import { FiHome, FiPackage } from "react-icons/fi";
import { GoBug } from "react-icons/go";
import { MdInfoOutline } from "react-icons/md";
import { PiLinkSimpleBold } from "react-icons/pi";
import { Link, useLocation } from "react-router-dom";
import { KURTOSIS_CLOUD_CONNECT_URL } from "../client/constants";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";

export const Navbar = () => {
  //const { updateSetting, settings } = useSettings();
  const location = useLocation();
  const kurtosisClient = useKurtosisClient();
  const [showAboutDialog, setShowAboutDialog] = useState(false);
  const kurtosisVersion = process.env.REACT_APP_VERSION || "Unknown";

  return (
    <Navigation>
      <Link to={"/"}>
        <NavButton
          label={"View enclaves"}
          Icon={<FiHome />}
          isActive={location.pathname === "/" || location.pathname.startsWith("/enclave")}
        />
      </Link>
      <Link to={"/catalog"}>
        <NavButton label={"View catalog"} Icon={<FiPackage />} isActive={location.pathname.startsWith("/catalog")} />
      </Link>
      {kurtosisClient.isRunningInCloud() && (
        <Link to={KURTOSIS_CLOUD_CONNECT_URL}>
          <NavButton label={"Link your CLI"} Icon={<PiLinkSimpleBold />} />
        </Link>
      )}
      <NavigationDivider />
      <Link
        to={`https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=&labels=bug&projects=&template=bug-report.yml&version=${kurtosisVersion}`}
        target={"_blank"}
      >
        <NavButton label={"Report a Bug"} Icon={<GoBug />} />
      </Link>
      <NavButton label={"About"} Icon={<MdInfoOutline />} onClick={() => setShowAboutDialog(true)} />

      <Modal isOpen={showAboutDialog} onClose={() => setShowAboutDialog(false)}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>About your Kurtosis Engine</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Text>Your Kurtosis engine version is:</Text>
            <InputGroup size={"lg"} variant={"solid"}>
              <Input
                value={kurtosisVersion}
                textOverflow={"ellipsis"}
                fontFamily={"Inconsolata"}
                bgColor={"gray.850"}
                readOnly
              />
              <InputRightElement>
                <CopyButton
                  contentName={"version"}
                  isIconButton
                  aria-label={"Click to copy this version"}
                  valueToCopy={kurtosisVersion}
                />
              </InputRightElement>
            </InputGroup>
            {/*<Text*/}
            {/*  as={"h2"}*/}
            {/*  fontSize={"lg"}*/}
            {/*  m={"16px 0"}*/}
            {/*  pt={"16px"}*/}
            {/*  borderTopWidth={"1px"}*/}
            {/*  borderTopColor={"gray.60"}*/}
            {/*  fontWeight={"semibold"}*/}
            {/*>*/}
            {/*  Settings:*/}
            {/*</Text>*/}
            {/*<FormControl display="flex" alignItems="center">*/}
            {/*  <FormLabel htmlFor="experimental-build" mb="0">*/}
            {/*    Enable experimental enclave builder interface?*/}
            {/*  </FormLabel>*/}
            {/*  <Switch*/}
            {/*    id="experimental-build"*/}
            {/*    onChange={() =>*/}
            {/*      updateSetting(*/}
            {/*        settingKeys.ENABLE_EXPERIMENTAL_BUILD_ENCLAVE,*/}
            {/*        !settings.ENABLE_EXPERIMENTAL_BUILD_ENCLAVE,*/}
            {/*      )*/}
            {/*    }*/}
            {/*    isChecked={settings.ENABLE_EXPERIMENTAL_BUILD_ENCLAVE}*/}
            {/*  />*/}
            {/*</FormControl>*/}
          </ModalBody>

          <ModalFooter>
            <Button onClick={() => setShowAboutDialog(false)}>Close</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Navigation>
  );
};

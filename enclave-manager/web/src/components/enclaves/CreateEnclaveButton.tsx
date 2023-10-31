import {Button, Menu, MenuButton, MenuItem, MenuList} from "@chakra-ui/react";
import {useCallback, useState} from "react";
import {FiPackage, FiPlus, FiSettings} from "react-icons/fi";
import {useNavigate} from "react-router-dom";
import {CreateEnclaveDialog} from "./CreateEnclaveDialog";

export const CreateEnclaveButton = () => {
    const navigate = useNavigate();
    const [manualCreateEnclaveOpen, setManualCreateEnclaveOpen] = useState(false);

    const handleManualCreateEnclaveClick = () => {
        setManualCreateEnclaveOpen(true);
    };

    const onDialogVisibilityChange = useCallback((isOpen: boolean) => {
        setManualCreateEnclaveOpen(isOpen);
    }, []);

    return (
        <>
            <Menu matchWidth>
                <MenuButton as={Button} colorScheme={"kurtosisGreen"} leftIcon={<FiPlus/>} size={"md"}>
                    Create Enclave
                </MenuButton>
                <MenuList>
                    <MenuItem onClick={handleManualCreateEnclaveClick} icon={<FiSettings/>}>
                        Manual
                    </MenuItem>
                    <MenuItem onClick={() => navigate("/catalog")} icon={<FiPackage/>}>
                        Catalog
                    </MenuItem>
                </MenuList>
            </Menu>
            <CreateEnclaveDialog
                defaultIsOpen={manualCreateEnclaveOpen}
                visibilityChangedCallback={onDialogVisibilityChange}
            />
        </>
    );
};

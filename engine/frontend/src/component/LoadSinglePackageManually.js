import {
    Button, Code,
    FormControl, FormErrorMessage, FormHelperText,
    FormLabel,
    Input,
    Modal,
    ModalBody,
    ModalCloseButton,
    ModalContent,
    ModalFooter,
    ModalHeader,
    ModalOverlay,
    useDisclosure
} from "@chakra-ui/react";
import {useEffect, useRef, useState} from "react";
import {useNavigate} from "react-router";
import {
    getOwnerNameFromUrl,
    getSinglePackageManually,
    getSinglePackageManuallyWithFullUrl
} from "../api/packageCatalog";

const LoadSinglePackageManually = () => {
    const {isOpen, onOpen, onClose} = useDisclosure()
    const [isLoading, setIsLoading] = useState(false)
    const navigate = useNavigate();
    const [input, setInput] = useState('')
    const handleInputChange = (e) => setInput(e.target.value)

    const initialRef = useRef(null)
    const finalRef = useRef(null)

    useEffect(() => {
        console.log(isOpen)
    }, [isOpen])

    useEffect(() => {
        console.log(input)
    }, [input])


    const isError = () => {
        try {
            getOwnerNameFromUrl(input)
        } catch (e) {
            console.log(e)
            return true
        }
        return false
    }

    const loadAndConfigure = () => {
        if (!isError()) {
            setIsLoading(true)
            getSinglePackageManuallyWithFullUrl(input)
                // getSinglePackageManuallyWithFullUrl("github.com/adschwartz/basic-service-package")
                // getSinglePackageManually(
                //     "github.com",
                //     "adschwartz",
                //     "basic-service-package",
                // )
                .then(r =>
                    navigate("/catalog/create", {state: {kurtosisPackage: r.package}})
                )
        }
    }

    return (
        <>
            <Button onClick={onOpen}>Open Modal</Button>
            {/*<Button ml={4} ref={finalRef}>*/}
            {/*    I'll receive focus on close*/}
            {/*</Button>*/}

            <Modal
                initialFocusRef={initialRef}
                finalFocusRef={finalRef}
                isOpen={isOpen}
                onClose={onClose}
            >
                <ModalOverlay/>
                <ModalContent>
                    <ModalHeader>Load custom package</ModalHeader>
                    <ModalCloseButton/>
                    <ModalBody pb={6}>
                        <FormControl isInvalid={isError()}>
                            <FormLabel>Enter Github URL</FormLabel>
                            <Input ref={initialRef}
                                   value={input} onChange={handleInputChange}
                                   placeholder=''/>
                            {!isError ? (
                                <FormHelperText>
                                    Enter the package you would like to load, e.g.
                                    `github.com/kurtosis-tech/etcd-package`
                                </FormHelperText>
                            ) : (
                                <FormErrorMessage>The package URL must follow the template
                                    'github.com/owner/repo'</FormErrorMessage>
                            )}
                        </FormControl>
                    </ModalBody>

                    <ModalFooter>
                        <Button variant='ghost' mr={3} onClick={onClose}>Cancel</Button>
                        <Button
                            type="submit"
                            colorScheme='green'
                            onClick={loadAndConfigure}
                            isDisabled={isLoading || isError()}
                            loadingText="Loading package..."
                        >
                            Load
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </>
    )
}

export default LoadSinglePackageManually;

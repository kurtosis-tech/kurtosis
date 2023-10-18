import {
    Button,
    FormControl,
    FormErrorMessage, FormHelperText,
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
import {useEffect, useState} from "react";
import {useNavigate} from "react-router";
import {getOwnerNameFromUrl, getSinglePackageManuallyWithFullUrl} from "../api/packageCatalog";

const LoadSinglePackageManually = () => {
    const {isOpen, onOpen, onClose} = useDisclosure()
    const [isLoading, setIsLoading] = useState(false)
    const navigate = useNavigate();
    const [input, setInput] = useState('')
    const [validUrl, setValidUrl] = useState(null)
    const [errorMessage, setErrorMessage] = useState("")
    const [modal, setModal] = useState(<></>)
    const handleInputChange = (e) => setInput(e.target.value)

    // github.com/adschwartz/text

    useEffect(() => {
        if (input === '') {
            setErrorMessage("")
        } else if (validUrl !== null && !validUrl) {
            setErrorMessage("No package was found at the entered URL")
        } else if (isError()) {
            setErrorMessage("The package URL must follow the template 'github.com/owner/repo'")
        }

    }, [validUrl, input])

    useEffect(() => {
        // clear the url validity when new data is entered
        setValidUrl(null)
    }, [input])

    const isError = () => {
        if(validUrl !== null && !validUrl) {
            return true
        }
        if (input === "") {
            return false
        }
        try {
            getOwnerNameFromUrl(input)
        } catch (e) {
            return true
        }
        return false
    }

    const loadAndConfigure = () => {
        if (!isError()) {
            setIsLoading(true)
            getSinglePackageManuallyWithFullUrl(input)
                .then(r => {
                        if (!r.package) {
                            setValidUrl(false)
                            setIsLoading(false)
                            setErrorMessage("No package was found at the entered URL")
                        } else {
                            navigate("/catalog/create", {state: {kurtosisPackage: r.package}})
                        }
                    }
                )
        }
    }

    const close = () => {
        onClose()
        navigate("/catalog")
    }

    useEffect(() => {
        setModal(
            <Modal
                isOpen={true}
                onClose={close}
            >
                <ModalOverlay/>
                <ModalContent>
                    <ModalHeader>Load Package</ModalHeader>
                    <ModalCloseButton/>
                    <ModalBody pb={6}>
                        <FormControl isInvalid={isError()}>
                            <FormLabel>Enter Github URL to package</FormLabel>
                            <Input value={input} onChange={handleInputChange} placeholder='github.com/owner/repo'/>
                            {/*<FormHelperText>Example with nested package:`github.com/kurtosis-tech/awesome-kurtosis/quickstart`</FormHelperText>*/}
                            <FormErrorMessage>{errorMessage}</FormErrorMessage>
                        </FormControl>
                    </ModalBody>
                    <ModalFooter>
                        <Button variant='ghost'
                                mr={3}
                                onClick={close}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            colorScheme='green'
                            onClick={loadAndConfigure}
                            isDisabled={isLoading || isError() || input === ""}
                            loadingText="Loading package..."
                        >
                            Load
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        )
    }, [errorMessage, input, isLoading, validUrl])

    return modal;
}

export default LoadSinglePackageManually;

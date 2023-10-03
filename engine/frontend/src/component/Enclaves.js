import {useNavigate} from "react-router-dom";
import NoData from "./NoData";
import LoadingOverlay from "./LoadingOverflow";
import {useRef, useState} from "react";
import {
    AlertDialog,
    AlertDialogBody,
    AlertDialogContent,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogOverlay,
    Button,
    Center,
    Grid,
    GridItem,
    Text,
    useDisclosure,
    FormControl,
    FormHelperText,
    Input, Tooltip, IconButton
} from '@chakra-ui/react'
import {DeleteIcon, EditIcon} from "@chakra-ui/icons";

const DeleteAlertDialog = ({isOpen, cancelRef, onClose, enclaveToDelete, setEnclave, handleDeleteClick}) => {
    const [deleting, setDeleting] = useState(false);
    const [value, setValue] = useState("")
    const [error, setError] = useState(false)

    const enclaveName = enclaveToDelete.name;
    const handleClose = (action) => {

        const clickCancel = () => {
            setError(false)
            setEnclave({})
            onClose()
        }

        const clickDelete = async () => {
            setDeleting(true)
            await handleDeleteClick(enclaveName)
            setDeleting(false)
            setError(false)
            clickCancel()
        }

        const maybeDeleteRequest = async () => {
            if (action === "delete") {
                if (value === enclaveName || !enclaveToDelete.mode) {
                    await clickDelete()
                } else {
                    setError(true)
                    setValue("")
                }
            } else {
                clickCancel()
            }
        }

        maybeDeleteRequest()
    }

    const handleInputChange = (val) => {
        if (error) {
            setError(false)
        }
        setValue(val)
    }

    return (
        <AlertDialog
            isOpen={isOpen}
            leastDestructiveRef={cancelRef}
            onClose={() => {
                setError(false)
                setValue("");
                onClose();
            }}
            isCentered
        >
            <AlertDialogOverlay>
                <AlertDialogContent backgroundColor={"white"}>
                    <AlertDialogHeader fontSize='lg' color={"black"}>
                        Delete Enclave: <Text fontSize='lg' fontWeight='bold' as='i'> {enclaveName} </Text>
                    </AlertDialogHeader>

                    <AlertDialogBody>
                        {
                            enclaveToDelete.mode ?
                                <FormControl>
                                    <Input onChange={(e) => handleInputChange(e.target.value)} isInvalid={error}
                                           borderColor={"black"} color={"black"}/>
                                    <FormHelperText color={error ? "red.600" : "black"} fontSize={"sm"}>
                                        {error ?
                                            "Please verify that the input matches the enclave name" :
                                            "Enter the enclave name to delete the enclave"
                                        }
                                    </FormHelperText>
                                </FormControl> :
                                <Text color={"black"}> Are you sure? You can't undo this action afterwards. </Text>
                        }

                    </AlertDialogBody>

                    <AlertDialogFooter>
                        <Button ref={cancelRef} onClick={() => handleClose("cancel")} color={"black"}>
                            Cancel
                        </Button>
                        <Button bg="red.600" _hover={{bg: "red.700"}} color="white"
                                onClick={() => handleClose("delete")} ml={3} isLoading={deleting}>
                            Delete
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialogOverlay>
        </AlertDialog>
    )
}

const Enclave = ({name, status, handleClick, onOpen, mode, setEnclave, handleEditEvent, host, port}) => {
    const backgroundColor = status === 1 ? "bg-[#24BA27]" : "bg-red-500"
    return (
        <Grid
            templateRows='repeat(3, 1fr)'
            templateColumns='repeat(1, 1fr)'
            className={`h-48 rounded-md border-4 ${backgroundColor} text-white items-center justify-center text-2xl`}
            onClick={() => handleClick(name)}
        >
            <GridItem colSpan={4} align={"right"} style={{"zIndex": 100}}>
                <Tooltip label='Edit a running enclave' fontSize='sm'>
                    <IconButton
                        boxSize={12}
                        mr="2"
                        colorScheme=''
                        _hover={{border: "1px"}}
                        aria-label='Edit Enclave'
                        icon={<EditIcon/>}
                        onClick={(e) => {
                            e.stopPropagation()
                            // setEnclave({
                            //     host: host,
                            //     port: port,
                            //     name: name,
                            // })
                            handleEditEvent(name, host, port)
                        }}
                    />
                </Tooltip>
                <Tooltip label='Delete an enclave' fontSize='sm'>
                    <IconButton
                        boxSize={12}
                        mr="2"
                        bg="red.600"
                        _hover={{bg: "red.700"}}
                        aria-label='Delete Enclave'
                        icon={<DeleteIcon/>}
                        onClick={(e) => {
                            e.stopPropagation()
                            setEnclave({mode: mode, name: name})
                            onOpen()
                        }}
                    />
                </Tooltip>
            </GridItem>
            <GridItem colSpan={4}>
                <Center>
                    <p className="text-3xl"> {name} </p>
                </Center>
            </GridItem>
            <GridItem colSpan={4} bg='papayawhip'>
            </GridItem>
        </Grid>
    )
}

const EnclaveMainComponent = ({onOpen, enclaves, handleClick, handleDeleteClick, setEnclave, handleEditEvent}) => (
    <div className='grid grid-cols-2 gap-4 flex-1'>
        {
            enclaves.map(enclave => {
                return (
                    <Enclave
                        onOpen={onOpen}
                        key={enclave.name}
                        name={enclave.name}
                        status={enclave.status}
                        created={enclave.created}
                        host={enclave.host}
                        port={enclave.port}
                        handleClick={handleClick}
                        handleDeleteClick={handleDeleteClick}
                        setEnclave={setEnclave}
                        mode={enclave.mode ? enclave.mode : false}
                        handleEditEvent={handleEditEvent}
                    />
                )
            })
        }
    </div>
)

const EnclaveComponent = ({
                              onOpen,
                              enclaves,
                              handleClick,
                              handleCreateEnvClick,
                              handleDeleteClick,
                              setEnclave,
                              handleEditEvent
                          }) => {
    return (
        <div className="flex-1 bg-[#171923] overflow-auto">
            {
                (enclaves.length === 0) ?
                    <div>
                        <NoData text={"No Enclaves Created"}/>
                        <div className="flex flex-row w-full">
                            <div className="w-1/3"></div>
                            <div className="mb-4 bg-[#24BA27] h-16 rounded w-1/3">
                                <div className='text-center w-full'>
                                    <div className='cursor-default text-3xl text-slate-800 p-2'
                                         onClick={handleCreateEnvClick}>
                                        Create Enclave
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    :
                    <EnclaveMainComponent setEnclave={setEnclave}
                                          onOpen={onOpen}
                                          enclaves={enclaves}
                                          handleClick={handleClick}
                                          handleDeleteClick={handleDeleteClick}
                                          handleEditEvent={handleEditEvent}
                    />
            }
        </div>
    )
}

const Enclaves = ({enclaves, isLoading, handleDeleteClick}) => {
    const navigate = useNavigate()
    const cancelRef = useRef()
    const [enclaveToDelete, setEnclave] = useState({})
    const {isOpen, onOpen, onClose} = useDisclosure()

    console.log("Enclaves", enclaves)


    const handleCreateEnvClick = () => {
        navigate("/catalog")
    }

    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }

    const handleEditEvent = (name, host, port) => {
        navigate(`/catalog/edit`,
            {
                state: {
                    name: name,
                    host: host,
                    port: port,
                }
            }
        )
    }

    return (
        <div className="flex h-full flex-grow">
            {
                (isLoading) ? <LoadingOverlay/> :
                    <EnclaveComponent setEnclave={setEnclave}
                                      onOpen={onOpen}
                                      enclaves={enclaves}
                                      handleClick={handleClick}
                                      handleCreateEnvClick={handleCreateEnvClick}
                                      handleDeleteClick={handleDeleteClick}
                                      handleEditEvent={handleEditEvent}
                    />
            }
            <DeleteAlertDialog
                isOpen={isOpen}
                onOpen={onOpen}
                onClose={onClose}
                cancelRef={cancelRef}
                enclaveToDelete={enclaveToDelete}
                setEnclave={setEnclave}
                handleDeleteClick={handleDeleteClick}
            />
        </div>
    )
}

export default Enclaves;

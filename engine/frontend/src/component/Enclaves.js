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
    useDisclosure
} from '@chakra-ui/react'

const DeleteAlertDialog = ({isOpen, cancelRef, onClose, enclaveName, setEnclaveName, handleDeleteClick}) => {
    const [deleting, setDeleting] = useState(false);

    const handleClose = (action) => {
        const maybeDeleteRequest = async (action) => {
            if (action === "delete") {
                setDeleting(true)
                await handleDeleteClick(enclaveName)
                setDeleting(false)
            }
            setEnclaveName("")
            onClose()
        }
        maybeDeleteRequest(action)
    }

    return (
        <AlertDialog
            isOpen={isOpen}
            leastDestructiveRef={cancelRef}
            onClose={onClose}
            isCentered
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize='lg'>
              Delete Enclave: <Text fontSize='lg' fontWeight='bold' as='i'> {enclaveName} </Text>
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure? You can't undo this action afterwards.
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={() => handleClose("cancel")}>
                Cancel
              </Button>
              <Button bg="red.600" _hover={{ bg: "red.700"}}  color="white" onClick={() => handleClose("delete")} ml={3} isLoading={deleting}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    )
}

const Enclave = ({name, status, created, handleClick, handleDeleteClick, onOpen, setEnclaveName}) => {
    const backgroundColor = status === 1 ? "bg-[#24BA27]" : "bg-red-500"
    return (
        <Grid
            templateRows='repeat(3, 1fr)'
            templateColumns='repeat(1, 1fr)'
            className={`h-48 rounded-md border-4 ${backgroundColor} text-white items-center justify-center text-2xl`}
            onClick={() => handleClick(name)}
        >
            <GridItem colSpan={4} align={"right"} style={{"zIndex":100}}>
                <Button bg="red.600" _hover={{ bg: "red.700"}} color="white" mr="2" onClick={(e)=> {
                    e.stopPropagation()
                    setEnclaveName(name)
                    onOpen()
                }}> Delete </Button>
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

const EnclaveMainComponent = ({onOpen, enclaves, handleClick, handleDeleteClick, setEnclaveName}) => (
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
                        handleClick={handleClick}
                        handleDeleteClick={handleDeleteClick}
                        setEnclaveName={setEnclaveName}
                    />
                )
            })
        }
    </div>
)

const EnclaveComponent = ({onOpen, enclaves, handleClick, handleCreateEnvClick, handleDeleteClick, setEnclaveName}) => {
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
                    <EnclaveMainComponent setEnclaveName={setEnclaveName} onOpen={onOpen} enclaves={enclaves} handleClick={handleClick} handleDeleteClick={handleDeleteClick}/>
            }
        </div>
    )
}

const Enclaves = ({enclaves, isLoading, handleDeleteClick}) => {
    const navigate = useNavigate()
    const cancelRef = useRef()
    const [enclaveName, setEnclaveName] = useState("")
    const { isOpen, onOpen, onClose } = useDisclosure()
    
    const handleCreateEnvClick = () => {
        navigate("/catalog")
    }

    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }

    return (
        <div className="flex h-full flex-grow">
            {
                (isLoading) ? <LoadingOverlay/> : <EnclaveComponent setEnclaveName={setEnclaveName} onOpen={onOpen} enclaves={enclaves} handleClick={handleClick} handleCreateEnvClick={handleCreateEnvClick} handleDeleteClick={handleDeleteClick}/>
            }
            <DeleteAlertDialog 
                isOpen={isOpen} 
                onOpen={onOpen} 
                onClose={onClose}
                cancelRef={cancelRef}
                enclaveName={enclaveName}
                setEnclaveName={setEnclaveName}
                handleDeleteClick={handleDeleteClick}
            />
        </div>
    ) 
}

export default Enclaves;

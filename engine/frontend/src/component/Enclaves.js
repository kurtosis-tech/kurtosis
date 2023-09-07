import {useNavigate} from "react-router-dom";
import NoData from "./NoData";
import LoadingOverlay from "./LoadingOverflow";
import {removeEnclave} from "../api/enclave";
import { Grid, GridItem, Center, Button, Tooltip } from '@chakra-ui/react'

const Enclave = ({name, status, created, handleClick, handleDeleteClick}) => {
    const backgroundColor = status === 1 ? "bg-[#24BA27]" : "bg-red-600"
    return (
        <Grid
            templateRows='repeat(3, 1fr)'
            templateColumns='repeat(1, 1fr)'
            className={`h-48 rounded-md border-4 ${backgroundColor} text-white items-center justify-center text-2xl`}
            onClick={() => handleClick(name)}
        >
            <GridItem colSpan={4} align={"right"} style={{"z-index":100}}>
                <Button colorScheme="red" color="white" mr="2" onClick={(e)=>{
                    e.stopPropagation()
                    handleDeleteClick(name)
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

const EnclaveMainComponent = ({enclaves, handleClick, handleDeleteClick}) => (
    <div className='grid grid-cols-2 gap-4 flex-1'>
        {
            enclaves.map(enclave => {
                return (
                    <Enclave
                        key={enclave.name}
                        name={enclave.name}
                        status={enclave.status}
                        created={enclave.created}
                        handleClick={handleClick}
                        handleDeleteClick={handleDeleteClick}
                    />
                )
            })
        }
    </div>
)

const EnclaveComponent = ({enclaves, handleClick, handleCreateEnvClick, handleDeleteClick}) => {
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
                    <EnclaveMainComponent enclaves={enclaves} handleClick={handleClick} handleDeleteClick={handleDeleteClick}/>
            }
        </div>
    )
}

const Enclaves = ({enclaves, isLoading, handleDeleteClick}) => {
    const navigate = useNavigate()

    const handleCreateEnvClick = () => {
        navigate("/catalog")
    }

    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }

    return (
        <div className="flex h-full flex-grow">
            {
                (isLoading) ? <LoadingOverlay/> : <EnclaveComponent enclaves={enclaves} handleClick={handleClick} handleCreateEnvClick={handleCreateEnvClick} handleDeleteClick={handleDeleteClick}/>
            }
        </div>
    ) 
}

export default Enclaves;

import RightPanel from "../RightPanel";
import LeftPanel from "../LeftPanel";
import {Log} from "./Log";
import {useNavigate} from "react-router-dom";
import {runStarlark} from "../../api/enclave";
import {getEnclaveInformation} from "../../api/container";
import LoadingOverlay from "../LoadingOverflow";
import {Box, Text, Flex, Spacer, Center} from "@chakra-ui/react";
import { useEffect, useState } from "react";

const SERVICE_IS_ADDED = "added with service";
const ERROR = "error"
const INSTRUCTION = "instruction"
const PROGRESS_INFO = "progressInfo"
const INSTRUCTION_RESULT = "instructionResult"


export const CreateEnclaveLog = ({packageId, enclave, args, appData}) => {
    console.log("enclave", enclave)

    const navigate = useNavigate();
    const [loading, setLoading] = useState(false)
    const [logs, setLogs] = useState([])
    const [services, setServices] = useState([])
    
    const getServices = async (enclave) => {
        const {services: newServices} = await getEnclaveInformation(enclave.host, enclave.port, appData.jwtToken, appData.apiHost);
        if (newServices.length > services.length) {
            setServices(newServices)
        }
    }

    const readStreamData = (result) => {
        if (result.case === INSTRUCTION) {
            setLogs(logs => [...logs, result.value.executableInstruction])
        }

        if (result.case === PROGRESS_INFO && result.value.currentStepInfo.length > 0) {
            if (result.value.currentStepInfo[result.value.currentStepNumber] !== undefined) {
                setLogs(logs => [...logs, result.value.currentStepInfo[result.value.currentStepNumber]])
            }
        }

        if (result.case === INSTRUCTION_RESULT && result.value.serializedInstructionResult) {
            if (result.value.serializedInstructionResult.includes(SERVICE_IS_ADDED)) {
                getServices(enclave)
            }
            setLogs(logs => [...logs, result.value.serializedInstructionResult])
        }

        if (result.case === ERROR) {
            const errorMessage = result.value.error.value.errorMessage;
            setLogs(logs => [...logs, errorMessage])
        }
    }

    useEffect(() => {
        setLoading(true)
        let stream;
        const fetchLogs = async () => {
            try {
                stream = await runStarlark(enclave.host, enclave.port, packageId, args, appData.jwtToken, appData.apiHost);
                for await (const res of stream) {
                    const result = res["runResponseLine"]
                    readStreamData(result)    
                }
            } catch (ex) {
                console.error("Error occurred while reading data from the enclave: ", enclave.name)
            } finally {
                setLoading(false)
            }
        }
        fetchLogs();
    }, [packageId])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclave.name}/services/${service.serviceUuid}`, {state: {services, selected: service}})
    }

    const renderServices = (services, handleClick) => {
        return services.map(service => {
            return (
                <div className={`flex items-center justify-center h-14 text-base bg-[#24BA27]`} key={service.name} onClick={()=>handleClick(service)}>
                    <div className='cursor-default text-lg text-white'> {service.name} </div>
                </div>
            )
        })
    }

    const renderLogView = () => {
        return (
            (logs.length > 0) ? <Log logs={logs} fileName={enclave.name} />: <Center color="white"> No Logs Available</Center>
        )
    }

    return (
        <div className="flex h-full">
            <div className="flex h-full">
                <LeftPanel 
                    home={false} 
                    heading={"Services"} 
                    isServiceInfo={true}
                    renderList={ ()=> renderServices(services, handleServiceClick)}
                />
                <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                    <div className='flex flex-col h-full space-y-1 bg-[#171923]'>
                        <Flex bg={"#171923"} height={`80px`}>    
                            <Box p='2' m="4"> 
                                <Text color={"white"} fontSize='xl' as='b'> 
                                    Logs  for {enclave.name} 
                                </Text>
                            </Box>
                            <Spacer/>
                        </Flex>
                        { (loading && logs.length === 0) ? <LoadingOverlay /> : renderLogView()}
                    </div>  
                </div>                    
                <RightPanel home={false} isServiceInfo={!loading} enclaveName={enclave.name}/>
            </div>
        </div>
    )
}

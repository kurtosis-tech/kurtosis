import { useEffect, useState } from "react";

import RightPanel from "./RightPanel";
import LeftPanel from "./LeftPanel";
import Heading  from "./Heading";
import { LogView } from "./LogView";

import {useNavigate} from "react-router-dom";
import {runStarlark} from "../api/enclave";
import {getEnclaveInformation} from "../api/container";
import LoadingOverlay from "./LoadingOverflow";

const SERVICE_IS_ADDED = "added with service";

export const CreateEnclaveView = ({packageId, enclave, args}) => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false)
    const [logs, setLogs] = useState([])
    const [services, setServices] = useState([])

    const getServices = async (apiClient) => {
        const {services: newServices} = await getEnclaveInformation(apiClient);
        if (newServices.length > services.length) {
            setServices(newServices) 
        }
    }

    useEffect(() => {
        setLoading(true)
        let stream;
        const fetchLogs = async () => {
          stream = await runStarlark(enclave.apiClient, packageId, args);
          stream.on("data", data => {
            const result = data.toObject();
            if (result.instruction && result.instruction.executableInstruction) {
                setLogs(logs => [...logs, result.instruction.executableInstruction])
            }

            if (result.progressInfo && result.progressInfo.currentStepInfoList.length > 0) {
                let length = result.progressInfo.currentStepInfoList.length;
                setLogs(logs => [...logs, result.progressInfo.currentStepInfoList[length-1]])
            } 
            
            if (result.instructionResult && result.instructionResult.serializedInstructionResult) {
                if (result.instructionResult.serializedInstructionResult.includes(SERVICE_IS_ADDED)) {
                    getServices(enclave.apiClient)
                }
                setLogs(logs => [...logs, result.instructionResult.serializedInstructionResult])
            }

            if (result.error) {
                if (result.error.interpretationError) {
                    setLogs(logs => [...logs, result.error.interpretationError.errorMessage])
                }

                if (result.error.executionError) {
                    setLogs(logs => [...logs, result.error.executionError.errorMessage])
                }

                if (result.error.validationError) {
                    setLogs(logs => [...logs, result.error.validationError.errorMessage])
                }
            }
          });

          stream.on("end", () => {
            setLoading(false)
          });
        }

        fetchLogs();
    }, [packageId])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclave.name}/services/${service.uuid}`, {state: {services, selected: service}})
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

    return (
        <div className="flex h-full bg-white">
            <div className="flex h-full">
                <LeftPanel 
                    home={false} 
                    heading={"Services"} 
                    isServiceInfo={true}
                    renderList={ ()=> renderServices(services, handleServiceClick)}
                />
                <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                    <div className='flex flex-col h-full space-y-1 bg-white'>
                        { (loading && logs.length === 0) ? <LoadingOverlay /> : <LogView 
                            heading={`Starlark Logs: ${enclave.name}`} 
                            logs={logs}
                            size={"h-full"}
                        />}
                    </div>  
                </div>                    
                <RightPanel home={false} isServiceInfo={!loading} enclaveName={enclave.name}/>
            </div>
        </div>
    )
}

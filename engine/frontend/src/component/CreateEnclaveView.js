import { useEffect, useState } from "react";

import RightPanel from "./RightPanel";
import LeftPanel from "./LeftPanel";
import Heading  from "./Heading";
import { LogView } from "./LogView";

import {useNavigate} from "react-router-dom";
import {runStarlark} from "../api/enclave";
import {getEnclaveInformation} from "../api/container";

const SERVICE_IS_ADDED = "added with service";

export const CreateEnclaveView = ({packageId, enclaveInfo, args}) => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false)
    const [logs, setLogs] = useState([])
    const [enclave, setEnclave] = useState("")
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
        const fetch = async () => {
          stream = await runStarlark(enclaveInfo.apiClient, packageId, args);
          setEnclave(enclave);
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
                    getServices(enclaveInfo.apiClient)
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

        fetch();
        return () => {
            if (stream) {
                stream.cancel();
            };
        };
    }, [packageId])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclaveInfo.enclave.name}/services/${service.uuid}`, {state: {services, selected: service}})
    }

    const renderServices = (services, handleClick) => {
        return services.map(service => {
            return (
                <div className={`flex items-center justify-center h-14 text-base bg-green-700`} key={service.name} onClick={()=>handleClick(service)}>
                    <div className='cursor-default text-lg text-white'> {service.name} </div>
                </div>
            )
        })
    }
    
    return (
        <div className="flex h-full w-full bg-white">
            <div className="flex h-full">
                <LeftPanel 
                    home={false} 
                    heading={"Services"} 
                    isServiceInfo={true}
                    renderList={ ()=> renderServices(services, handleServiceClick)}
                />
                <div className="flex h-full w-[calc(100vw-24rem)] flex-col space-y-5">
                    <div className='flex flex-col h-full space-y-1 bg-white'>
                        <LogView 
                            heading={`Starlark Logs: ${enclaveInfo.enclave.name}`} 
                            logs={logs}
                        />
                    </div>  
                </div>                    
                <RightPanel home={false} isServiceInfo={true} enclaveName={enclaveInfo.enclave.name}/>
            </div>
        </div>
    )
}
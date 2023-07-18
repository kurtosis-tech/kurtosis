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
        const {services} = await getEnclaveInformation(apiClient);
        setServices(services) 
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
            <LeftPanel 
                home={false} 
                heading={"Services"} 
                renderList={ ()=> renderServices(services, handleServiceClick)}
            />
            <div className="flex-1">
                <Heading content={enclaveInfo.enclave.name} color={"text-black"}/>
                <LogView classAttr={"flex flex-col p-2"} heading={"Logs"} logs={logs} loading={loading && logs.length===0}/>
            </div>
                    
            <RightPanel home={false}/>
        </div>
    )
}
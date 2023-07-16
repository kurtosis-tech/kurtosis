import { useEffect, useState } from "react";

import RightPanel from "../component/RightPanel";
import LeftPanel from "../component/LeftPanel";
import Heading  from "../component/Heading";
import { LogView } from "./LogView";

import {useNavigate} from "react-router-dom";

import {runStarlark} from "../api/enclave";
import {getEnclaveInformation} from "../api/container";

const SERVICE_IS_ADDED = "added with service";

export const CreateEnclaveView = ({packageId, enclaveInfo}) => {
    const navigate = useNavigate();
    const [logs, setLogs] = useState([])
    const [enclave, setEnclave] = useState("")
    const [services, setServices] = useState([])

    const getServices = async (apiClient) => {
        const {services} = await getEnclaveInformation(apiClient);
        setServices(services) 
    }

    useEffect(() => {
        let stream;
        const fetch = async () => {
          stream = await runStarlark(enclaveInfo.apiClient, packageId);
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
            //navigate("/");
          });
        }

        fetch();
        return () => {
            if (stream) {
                stream.cancel();
            };
        };
    }, [packageId])

    const renderServices = () => {
        return services.map(service => {
            return (
                <div 
                    className={`cursor-default flex text-white items-center justify-center h-14 rounded-md border-4 bg-green-700`} 
                    key={service.uuid}>
                    {service.name}
                </div>
            )   
        }) 
    }

    return (
        <div className="flex h-full w-full bg-white">
            <LeftPanel 
                home={false} 
                heading={"Services"} 
                renderList={ ()=> renderServices(services, ()=>{})}
            />
            <div className="flex-1">
                <Heading content={enclaveInfo.enclave.name} color={"text-black"}/>
                <div className='flex flex-col h-full space-y-1'>                        
                    <LogView classAttr={"flex flex-col p-2"} heading={"Logs"} logs={logs}/>
                </div>
            </div>
                    
            <RightPanel home={false}/>
        </div>
    )
}
import { useEffect, useState } from "react";
import { LeftView } from "./LeftView";
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
            console.log(result);
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
        <div className="grid grid-cols-6 h-full w-full">
            <LeftView 
                heading={"Services"} 
                renderList={() => renderServices()}
            /> 
            <div className="col-span-5">
                <div className='flex flex-col h-screen space-y-5'>
                    <div className="text-3xl text-center"> 
                        {enclaveInfo.enclave.name} 
                    </div>
                    <LogView classAttr={"flex flex-col p-2 h-fit"} heading={"Starlark Logs"} logs={logs}/>
                </div>
            </div> 
        </div>
    )
}
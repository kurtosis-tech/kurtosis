import { useEffect, useState } from "react";

import RightPanel from "./RightPanel";
import LeftPanel from "./LeftPanel";
import { LogView } from "./LogView";

import {useLocation, useNavigate} from "react-router-dom";
import {runStarlark} from "../api/enclave";
import {getEnclaveInformation} from "../api/container";
import LoadingOverlay from "./LoadingOverflow";

const SERVICE_IS_ADDED = "added with service";

const PackageCatalogProgress = ({appData}) => {
    const location = useLocation()
    const {state} = location;
    const enclave = state.enclave;
    const packageId = state.runArgs.packageId;
    const args = state.runArgs.args;

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

    useEffect(() => {
        setLoading(true)
        let stream;
        const fetchLogs = async () => {
            stream = await runStarlark(enclave.host, enclave.port, packageId, args, appData.jwtToken, appData.apiHost);
            for await (const res of stream) {
              const result = res["runResponseLine"]
              if (result.case === "instruction") {
                  setLogs(logs => [...logs, result.value.executableInstruction])
              }

              if (result.case === "progressInfo" && result.value.currentStepInfo.length > 0) {
                  console.log("progressinfo: ", result)
                  setLogs(logs => [...logs, result.value.currentStepInfo[result.value.currentStepNumber]])
              }

              if (result.case === "instructionResult" && result.value.serializedInstructionResult) {
                  if (result.value.serializedInstructionResult.includes(SERVICE_IS_ADDED)) {
                      getServices(enclave)
                  }
                  setLogs(logs => [...logs, result.value.serializedInstructionResult])
              }
            }
            setLoading(false)
        }
        fetchLogs();
    }, [packageId])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclave.name}/services/${service.uuid}`, {state: {services, selected: service}})
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

export default PackageCatalogProgress;
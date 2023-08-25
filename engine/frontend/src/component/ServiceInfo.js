import Heading from "./Heading";
import { useEffect, useState } from "react";
import {useNavigate, useParams, useLocation} from "react-router-dom";
import { LogView } from "./LogView";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import {getServiceLogs} from "../api/enclave";
import {useAppContext} from "../context/AppState";


const renderServices = (services, handleClick) => {
    return services.map(service => {
        return (
            <div className={`flex items-center justify-center h-14 text-base bg-[#24BA27]`} key={service.name} onClick={()=>handleClick(service)}>
                <div className='cursor-default text-lg text-white'> {service.name} </div>
            </div>
        )
    })
}

const renderFileArtifacts = (file_artifacts) => {
    if (file_artifacts.length === 0) {
        return (
            <div className="text-3xl text-slate-200 text-center justify-center">
                No Data
             </div>
        )
    }

    return file_artifacts.map((file_artifact)=> {
        return (
            <div className="border-4 bg-[#171923] text-lg align-middle text-center h-16 p-3 text-[#24BA27]">
                <div> {file_artifact.name} </div>
            </div>
        )
    })
}

const ServiceInfo = () => {
    const navigate = useNavigate();
    const [logs, setLogs] = useState([])
    const {state} = useLocation();
    const {services, selected} =  state;
    const {appData} = useAppContext()

    const params = useParams()
    const {name: enclaveName, uuid:serviceUuid} = params;

    useEffect(() => {
        let stream;
        const ctrl = new AbortController();
        const fetchLogs = async () => {
            stream = await getServiceLogs(ctrl, enclaveName, serviceUuid, appData.apiHost);
            try {
                for await (const res of stream) {
                    const log = res["serviceLogsByServiceUuid"][serviceUuid]["line"][0]
                    setLogs(logs => [...logs, log])
                }
            } catch (ex) {
                console.log("aborted stream!")
            }
        }
        fetchLogs()
        return () => {
            ctrl.abort()
            setLogs([]);
        };
    }, [serviceUuid])

    const handleServiceClick = (service) => {
        const fullPath = `../../../enclaves/${enclaveName}/services/${service.uuid}`
        navigate(fullPath, {state: {services, selected: service}})
    }

    return (
        <div className="flex h-full">
            <LeftPanel 
                home={false} 
                isServiceInfo={true}
                heading={"Services"} 
                renderList={ ()=> renderServices(services, handleServiceClick)}
            />
            <div className="flex h-full w-[calc(100vw-39rem)] flex-col space-y-5">
                <div className='flex flex-col h-full space-y-1 bg-white'>
                    <Heading content={`${enclaveName}::${selected.name}`} />
                    <div className="flex-1">
                        <div className="text-xl text-left h-fit mb-2 ml-5"> 
                            Ports
                        </div>
                        <div className="overflow-auto">
                                {
                                    selected.ports.map(port => {
                                        const urlWithApplicationString = `${port.applicationProtocol}://localhost:${port.publicPortNumber}`
                                        const urlWithoutApplicationString = `localhost:${port.publicPortNumber}`
                                        const url = port.applicationProtocol ? urlWithApplicationString: urlWithoutApplicationString 
                                        
                                        return (
                                            <div className="h-fit flex flex-row space-x-10 ml-5">
                                                <div> {port.portName}: </div> 
                                                <a href={url} rel="noreferrer" className="grow">
                                                    <u> {url} </u>
                                                </a>
                                            </div>  
                                        )
                                    })
                                }
                        </div>
                    </div>
                    <LogView heading={`Service Logs`} logs={logs}/>
                </div>  
            </div>                    
            <RightPanel home={false} isServiceInfo={true} enclaveName={enclaveName}/>
        </div>
    )
}

export default ServiceInfo;

{/* <div className="flex w-[calc(100vw-24rem)] flex-col bg-white">
                <Heading content={selected.name} color={"text-black"}/>
                <div className='flex flex-col h-full space-y-1'>
                    <div className="flex flex-col h-1/6 border-4 bg-green-200">
                        <div className="text-xl text-left h-fit mb-2 ml-5"> 
                            Ports
                        </div>
                        <div className="overflow-auto">
                                {
                                    selected.ports.map(port => {
                                        const urlWithApplicationString = `${port.applicationProtocol}://localhost:${port.publicPortNumber}`
                                        const urlWithoutApplicationString = `localhost:${port.publicPortNumber}`
                                        const url = port.applicationProtocol ? urlWithApplicationString: urlWithoutApplicationString 
                                        
                                        return (
                                            <div className="h-fit flex flex-row space-x-10 ml-5">
                                                <div> {port.portName}: </div> 
                                                <a href={url} rel="noreferrer" className="grow">
                                                    <u> {url} </u>
                                                </a>
                                            </div>  
                                        )
                                    })
                                }
                        </div>
                </div>
                </div>
            </div> */}

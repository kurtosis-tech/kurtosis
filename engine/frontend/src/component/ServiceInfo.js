import Heading from "./Heading";
import { useEffect, useState } from "react";
import {useNavigate, useParams, useLocation} from "react-router-dom";
import {getEnclaveInformation} from "../api/container";
import { LogView } from "../components/LogView";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import {getServiceLogs} from "../api/enclave";


const renderServices = (services, handleClick) => {
    if (services.length === 0) {
        return (
            <div className="text-3xl text-red-600 text-center justify-center">
                No Data: 
                This occurs because either enclave is stopped or there was error while executing
                the package.
             </div>
        )
    }

    return services.map(service => {
        return (
            <div className={`flex items-center justify-center h-14 text-base bg-green-700`} key={service.name} onClick={()=>handleClick(service)}>
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
            <div className="border-4 bg-slate-800 text-lg align-middle text-center h-16 p-3 text-green-600"> 
                <div> {file_artifact.name} </div>
            </div>
        )
    })
}

const ServiceInfo = ({enclaves}) => {
    const navigate = useNavigate();
    const [logs, setLogs] = useState([])
    const {state} = useLocation();
    const {services, selected} =  state;

    const params = useParams()
    const {name: enclaveName, uuid:serviceUuid} = params;

    console.log(selected)

    useEffect(() => {
        let stream;
        const fetch = async () => {
            stream = await getServiceLogs(enclaveName, serviceUuid);
            stream.on("data", data => {
                const log = data.toObject().serviceLogsByServiceUuidMap[0][1].lineList
                setLogs(logs => [...logs, log[0]])
            })
        }
        fetch()
        return () => {
            if (stream) {
                stream.cancel();
                // need to do this - this means that we are getting logs from
                // different service
                setLogs([])
            };
        };
    }, [serviceUuid])

    const handleServiceClick = (service) => {
        navigate(`/enclaves/${enclaveName}/services/${service.uuid}`, {state: {services, selected: service}})
    }

    // const handleLeftPanelClick = (enclaveName) => {
    //     navigate(`/enclaves/${enclaveName}`, {replace:true})
    // }

    return (
        <div className="flex h-full bg-white">
            <LeftPanel 
                home={false} 
                heading={"Services"} 
                renderList={ ()=> renderServices(services, handleServiceClick)}
            />

            <div className="flex-1">
                <Heading content={selected.name} color={"text-black"}/>
                <div className='flex flex-col h-full space-y-1'>
                    <div className="flex flex-col h-1/6 border-4">
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
                    <LogView classAttr={"flex flex-col p-2 h-5/6"} heading={"Service Logs"} logs={logs}/>
                </div>
            </div>
                    
            <RightPanel home={false}/>
        </div>
    )
}

export default ServiceInfo;
import { useEffect, useState } from "react";
import {useNavigate, useParams, useLocation} from "react-router-dom";
import {getServiceLogs} from "../api/enclave";
import { LeftView } from "./LeftView";
import { LogView } from "./LogView";

const ServiceInfo = () => {
    const navigate = useNavigate();
    const [logs, setLogs] = useState([])
    const {state} = useLocation();
    const {services, selected} =  state;

    const params = useParams()
    const {name: enclaveName, uuid:serviceUuid} = params;

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
            };
        };
    }, [serviceUuid])

    const handleLeftPanelClick = (service) => {
        navigate(`/enclaves/${enclaveName}/services/${service.uuid}`, {state: {services, selected: service}, replace:true})
    }

    const renderServices = (services, handleClick) => {
        return services.map(service => {
            return (
                <div 
                    className={`cursor-default flex text-white items-center justify-center h-14 rounded-md border-4 bg-green-700`} 
                    key={service.uuid} onClick={()=>handleClick(service)}>
                    {service.name}
                </div>
            )   
        }) 
    }

    return (
        <div className="grid grid-cols-6 h-full w-full">
            <LeftView 
                heading={enclaveName} 
                renderList={() => renderServices(services, handleLeftPanelClick)}
            /> 
            <div className="col-span-5">
                <div className='flex flex-col h-screen space-y-5'>
                    <div className="mt-3 mb-3 h-1/3 flex-col space-y-5"> 
                        <div className="text-3xl text-center"> 
                            {selected.name} 
                        </div>
                        <div className="flex flex-col">
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
                    <LogView classAttr={"flex flex-col p-2 h-2/3"} heading={"Service Logs"} logs={logs}/>
                </div>
            </div> 
        </div>
    )
}

export default ServiceInfo; 
import { useEffect, useState } from "react";
import {useNavigate, useParams, useLocation} from "react-router-dom";
import {getEnclaveInformation} from "../api/container";

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

    return services.map((service)=> {
        return (
            <div className="border-4 bg-slate-800 text-lg align-middle text-center h-16 p-3 text-green-600" onClick={() => handleClick(service, services)}> 
                <div> {service.name} </div>
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

const LeftPanel = ({enclaves, handleClick}) => {

    const renderEnclaves = (enclaves, handleClick) => {
        return enclaves.map(enclave => {
            const backgroundColor = enclave.status === 1 ? "bg-green-700": "bg-red-600"
            return (
                <div className={`cursor-default flex text-white items-center justify-center h-14 rounded-md border-4 ${backgroundColor}`} key={enclave.name} onClick={()=>handleClick(enclave.name)}>
                    {enclave.name}
                </div>
            )   
        }) 
    }

    return (
        <div className='flex flex-col h-screen space-y-5 bg-slate-800'>
            <div className="text-3xl text-center mt-5 mb-3 text-white"> 
                Environments 
            </div>
            <div className="flex flex-col space-y-4 p-2 overflow-auto">
                 {renderEnclaves(enclaves, handleClick)}
            </div> 
            {/* <div className="flex text-3xl justify-center py-10 text-white border-t-4"> {username} </div> */}
     </div>
    )
}

const Services = () => {
    const navigate = useNavigate();
    
    const params = useParams();
    const {name} = params;

    const [services, setServices] = useState([])
    const [fileArtifacts, setFileArtifacts] = useState([])

    const {state} = useLocation();
    const {enclaves} =  state;
    
    useEffect(() => {
        const fetch = async () => {
            //const response = await axios.get(`http://localhost:5050/enclaves/${name}/services`)
            //setFileArtifacts(response.data.body.fileArtifacts)
            const selected = enclaves.filter(enclave => enclave.name === name);
            const {services, artifacts} = await getEnclaveInformation(selected[0].apiClient);
            setServices(services)
            setFileArtifacts(artifacts)
        } 
        fetch()
    }, [name])

    const handleLeftPanelClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`, {state: {enclaves}, replace:true})
    }

    const handleServiceClick = (service, services) => {
        navigate(`/enclaves/${name}/services/${service.uuid}`, {state: {services, selected: service}})
    }

    return (
        <div className="grid grid-cols-6 h-full w-full">
            <div className="bg-slate-800"> 
                <LeftPanel navigate={navigate} enclaves={enclaves} handleClick={handleLeftPanelClick}/> 
            </div>
            <div className="col-span-5 gap-5 h-full space-y-5">
                <div className="flex flex-col space-y-2 overflow-auto text-4xl text-center justify-center pt-2">
                    {name}
                </div>
                <div className="flex flex-col space-y-2 h-1/2 border-8">
                    <div className="space-y-2 text-2xl text-center justify-center pt-2">
                        Services
                    </div>
                    <div className="overflow-auto">
                        {renderServices(services, handleServiceClick)}
                    </div>
                </div>  
                <div className="flex flex-col space-y-2 overflow-auto h-1/3 border-8">
                    <div className="flex flex-col space-y-2 overflow-auto text-2xl text-center justify-center pt-2">
                        Files Artifacts
                    </div>
                    <div className="overflow-auto">
                        {renderFileArtifacts(fileArtifacts)}
                    </div>
                </div>  
            </div>
            {/* <div className="col-span-5 grid grid-rows-6">
                <div className="text-center justify-center pt-5 text-2xl"> {name} </div>
                <div className="row-span-5"> 
                    <div className="text-center justify-center text-xl"> Services </div>
                    <div className="flex flex-col h-screen space-y-2">
                        {renderServices(services)}
                    </div>
                </div>
            </div> */}
            {/* {
                enclaves.map(enclave => {
                    return (
                       <Enclave key={enclave.name} name={enclave.name} status={enclave.status} created={enclave.created} />
                    )
                })
            } */}
        </div>
    )
}

export default Services; 
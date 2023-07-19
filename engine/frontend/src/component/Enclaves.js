import {useNavigate} from "react-router-dom";
import NoData from "./NoData";
import LoadingOverlay from "./LoadingOverflow";
import RightPanel from "./RightPanel";

const Enclave = ({name, status, created, handleClick}) => {
    const backgroundColor = status === 1 ? "bg-green-700": "bg-red-600"
    return (
        <div onClick={() => handleClick(name)}className={`h-48 p-4 rounded-md border-4 flex ${backgroundColor} text-white items-center justify-center text-2xl flex-col`}>
            <p className="text-3xl"> {name} </p>        
            <p className="text-xs"> {created} </p>
        </div>
    )
}

const EnclaveMainComponent = ({enclaves, handleClick}) => (
    <div className='grid grid-cols-2 gap-4 flex-1'>
     {
        enclaves.map(enclave => {
            return (
                <Enclave 
                    key={enclave.name} 
                    name={enclave.name} 
                    status={enclave.status} 
                    created={enclave.created}
                    handleClick={handleClick}
                />
                )
            })
   }       
    </div>
)

const EnclaveComponent = ({enclaves, handleClick}) => {
    return (
        <div className="flex-1 bg-slate-800 overflow-auto">
            {
                (enclaves.length === 0) ? <NoData /> : <EnclaveMainComponent enclaves={enclaves} handleClick={handleClick} />
            }
        </div>
    )
}

const Enclaves = ({enclaves, isLoading}) => {
    const navigate = useNavigate()
    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }
    return (
        <div className="flex h-full flex-grow">
            {
                (isLoading) ? <LoadingOverlay/> : <EnclaveComponent enclaves={enclaves} handleClick={handleClick}/>
            }
        </div>
    ) 
}

export default Enclaves;
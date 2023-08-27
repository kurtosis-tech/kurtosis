import {useNavigate} from "react-router-dom";
import NoData from "./NoData";
import LoadingOverlay from "./LoadingOverflow";

const Enclave = ({name, status, created, handleClick}) => {
    const backgroundColor = status === 1 ? "bg-[#24BA27]" : "bg-red-600"
    return (
        <div onClick={() => handleClick(name)}
             className={`h-48 p-4 rounded-md border-4 flex ${backgroundColor} text-white items-center justify-center text-2xl flex-col`}>
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

const EnclaveComponent = ({enclaves, handleClick, handleCreateEnvClick}) => {
    return (
        <div className="flex-1 bg-[#171923] overflow-auto">
            {
                (enclaves.length === 0) ?
                    <div>
                        <NoData text={"No Enclaves Created"}/>
                        <div className="flex flex-row w-full">
                            <div className="w-1/3"></div>
                            <div className="mb-4 bg-[#24BA27] h-16 rounded w-1/3">
                                <div className='text-center w-full'>
                                    <div className='cursor-default text-3xl text-slate-800 p-2'
                                         onClick={handleCreateEnvClick}>
                                        Create Enclave
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    :
                    <EnclaveMainComponent enclaves={enclaves} handleClick={handleClick}/>
            }
        </div>
    )
}

const Enclaves = ({enclaves, isLoading}) => {
    const navigate = useNavigate()

    const handleCreateEnvClick = () => {
        navigate("/enclave/create")
    }
    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }
    console.log("is loading:", isLoading)
    console.log("Updating with enclaves:", JSON.stringify(enclaves))
    return (
        <div className="flex h-full flex-grow">
            {
                (isLoading) ? <LoadingOverlay/> : <EnclaveComponent enclaves={enclaves} handleClick={handleClick} handleCreateEnvClick={handleCreateEnvClick}/>
            }
        </div>
    ) 
}

export default Enclaves;

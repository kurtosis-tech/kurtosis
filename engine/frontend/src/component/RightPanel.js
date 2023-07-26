import Heading  from "./Heading";
import { useNavigate } from "react-router";

const RightPanel = ({isServiceInfo, enclaveName}) => {
    const navigate = useNavigate()
    
    const handleGotoMenu = () => {
        navigate("/enclaves")
    }
    const handleGoToEnclave = () => {
        navigate(`/enclaves/${enclaveName}`)
    }
    const handleCreateEnclave = () => {
        navigate("/enclave/create")
    }

    return (
        <div className="flex-none w-[15rem] p-4 border-l bg-slate-800">
            <div className="flex flex-col space-y-10 items-center">
                {
                    isServiceInfo ? <button className="w-full bg-green-600 text-slate-800 h-14" onClick={handleGoToEnclave}> {`View ${enclaveName}`}  </button> : null
                }
                <button className="w-full bg-green-600 text-slate-800 h-14" onClick={handleGotoMenu}> Home </button>
                
                <svg class="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 18 18" onClick={handleCreateEnclave}>
                    <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 1v16M1 9h16"/>
                </svg>
            </div>
        </div>  
    ) 
}

export default RightPanel;
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
        navigate("/catalog")
    }

    return (
        <div className="flex-none w-[15rem] p-4 bg-[#171923]">
            <div className="flex flex-col space-y-10 items-center">
                {
                    isServiceInfo ? <button className="w-full bg-[#24BA27] text-slate-800 h-14" onClick={handleGoToEnclave}> {`View ${enclaveName}`}  </button> : null
                }
                <button className="w-full bg-[#24BA27] text-slate-800 h-14" onClick={handleGotoMenu}> Home </button>
                
                <svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 18 18" onClick={handleCreateEnclave}>
                    <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 1v16M1 9h16"/>
                </svg>
            </div>
        </div>  
    ) 
}

export default RightPanel;

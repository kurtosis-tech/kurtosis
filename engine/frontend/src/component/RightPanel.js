import Heading  from "./Heading";
import { useNavigate } from "react-router";

const RightPanel = () => {
    const navigate = useNavigate()
    
    const handleGotoMenu = () => {
        navigate("/enclaves")
    }

    const handleCreateEnclave = () => {
        navigate("/enclave/create")
    }

    return (
        <div className="flex-none w-fit p-4 border-l bg-slate-800">
            <div className="flex flex-col space-y-10 items-center">
                <svg class="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20" onClick={handleGotoMenu}>
                    <path d="m19.707 9.293-2-2-7-7a1 1 0 0 0-1.414 0l-7 7-2 2a1 1 0 0 0 1.414 1.414L2 10.414V18a2 2 0 0 0 2 2h3a1 1 0 0 0 1-1v-4a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v4a1 1 0 0 0 1 1h3a2 2 0 0 0 2-2v-7.586l.293.293a1 1 0 0 0 1.414-1.414Z"/>
                </svg>
                <svg class="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 18 18" onClick={handleCreateEnclave}>
                    <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 1v16M1 9h16"/>
                </svg>
            </div>
        </div>  
    ) 
}

export default RightPanel;
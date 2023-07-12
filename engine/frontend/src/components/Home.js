import { useEffect, useState } from "react";
import axios from 'axios';
import {useNavigate} from "react-router-dom";
import {getEnclavesFromKurtosis} from "../api/enclave";

const Enclave = ({name, status, created, handleClick}) => {
    const backgroundColor = status === 1 ? "bg-green-700": "bg-red-600"
    return (
        <div onClick={() => handleClick(name)}className={`h-48 rounded-md border-4 flex ${backgroundColor} text-white items-center justify-center text-2xl flex-col`}>
            <p className="text-3xl"> {name} </p>        
            <p className="text-xs"> {created} </p>
        </div>
    )
}

const Home = () => {
    const navigate = useNavigate()
    const [enclaves, setEnclaves] = useState([])

    useEffect(() => {
        const fetch = async () => {
            const response = await getEnclavesFromKurtosis();
            setEnclaves(enclaves => [...enclaves, ...response])
        } 
        fetch()
    }, [])

    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`, {state: {enclaves}}); // Navigate to the specified route
    }

    return (
        <div className="grid grid-cols-2 gap-4 h-96 overflow-auto h-full bg-slate-800">
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
}

export default Home; 
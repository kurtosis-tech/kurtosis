import Heading from "./Heading";
import {useNavigate} from "react-router-dom";
import {getEnclavesFromKurtosis} from "../api/enclave";
import { useEffect, useState } from "react";

const Enclave = ({name, status, created, handleClick}) => {
    const backgroundColor = status === 1 ? "bg-green-700": "bg-red-600"
    return (
        <div onClick={() => handleClick(name)}className={`h-48 p-4 rounded-md border-4 flex ${backgroundColor} text-white items-center justify-center text-2xl flex-col`}>
            <p className="text-3xl"> {name} </p>        
            <p className="text-xs"> {created} </p>
        </div>
    )
}


const Enclaves = ({enclaves}) => {
    const navigate = useNavigate()
    const handleClick = (enclaveName) => {
        navigate(`/enclaves/${enclaveName}`)
    }
    return (
        <div className="flex h-full">
            <div className="flex-1 bg-slate-800 overflow-auto">
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
            </div>
        </div>
    ) 
}

export default Enclaves;
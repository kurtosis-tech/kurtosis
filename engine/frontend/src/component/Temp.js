import React from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import CreateEnclave from '../components/CreateEnclave';
import Enclaves from "./Enclaves";
import { useEffect, useState } from "react";
import {getEnclavesFromKurtosis} from "../api/enclave";

import { BrowserRouter as Router, Route, Routes,useLocation, useNavigate, useParams } from 'react-router-dom';

const ENCLAVES = [
    {
        name: "one",
        status: 1,
    },
    {
        name: "two",
        status: 1,
    },
    {
        name: "three",
        status: 1,
    },
    {
        name: "four",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
    {
        name: "five",
        status: 1,
    },
]

const renderEnclaves = (enclaves, handleClick) => {
    return enclaves.map(enclave => {
        const backgroundColor = enclave.status === 1 ? "bg-green-700": "bg-red-600"
        return (
            <div className={`flex items-center justify-center h-14 text-base ${backgroundColor}`} key={enclave.name} onClick={()=>handleClick(enclave.name)}>
                <div className='cursor-default text-lg text-white'> {enclave.name} </div>
            </div>
        )
    })
}

const Temp = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const [enclaves, setEnclaves] = useState([])

    useEffect(() => {
        const fetch = async () => {
            const response = await getEnclavesFromKurtosis();
            setEnclaves(response)
        } 
        fetch()
    }, [])

    return (
        <div className="h-screen flex flex-col bg-slate-800">
            <TitleBar />
            {/* Main Component */}
            <div className="flex-grow overflow-hidden">
                <Routes>
                    <Route exact path="/" element={<Main/>} />
                    <Route exact path="/enclaves" element={<Enclaves enclaves={enclaves}/>} />
                    <Route exact path="/enclaves/create" element={<CreateEnclave />} />
                    <Route path="/enclaves/:name" element={<EnclaveInfo enclaves={enclaves}/>} />
                    <Route path="/enclaves/:name/services/:uuid" element={<ServiceInfo/>} />
                </Routes>
            </div>
        </div>
  );
}

export default Temp;

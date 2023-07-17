import React from 'react';
import TitleBar from "./TitleBar"
import Main from "./Main"
import EnclaveInfo from "./EnclaveInfo";
import ServiceInfo from "./ServiceInfo";
import CreateEnclave from './CreateEnclave';
import Enclaves from "./Enclaves";
import { useEffect, useState } from "react";
import {getEnclavesFromKurtosis} from "../api/enclave";

import { Route, Routes } from 'react-router-dom';

const Temp = () => {
    const [enclaves, setEnclaves] = useState([])
    const [encalveLoading, setEnclaveLoading] = useState(false)

    useEffect(() => {
        setEnclaveLoading(true)
        const fetch = async () => {
            const response = await getEnclavesFromKurtosis();
            setEnclaves(response)
            setEnclaveLoading(false)
        } 
        fetch()
    }, [])

    return (
        <div className="h-screen flex flex-col bg-slate-800">
            <TitleBar />
            <div className="flex-grow overflow-hidden">
                <Routes>
                    <Route exact path="/" element={<Main/>} />
                    <Route exact path="/enclaves" element={<Enclaves enclaves={enclaves} isLoading={encalveLoading}/>} />
                    <Route exact path="/enclaves/create" element={<CreateEnclave />} />
                    <Route path="/enclaves/:name" element={<EnclaveInfo enclaves={enclaves}/>} />
                    <Route path="/enclaves/:name/services/:uuid" element={<ServiceInfo/>} />
                </Routes>
            </div>
        </div>
  );
}

export default Temp;

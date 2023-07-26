import React, { useState} from 'react';

import {CreateEnclaveModal} from "./CreateEnclaveModal";
import {CreateEnclaveView} from "./CreateEnclaveView";
import { Route, Routes, useNavigate } from 'react-router-dom';

const CreateEnclave = ({addEnclave}) => {
    const navigate = useNavigate()
    const [name, setName] = useState('');
    const [args, setArgs] = useState('{}')
    const [enclave, setEnlave] = useState(null);

    const handleModalSubmit = (enclave) => {
        setEnlave(enclave)
        navigate("/enclave/progress")
    }

    return (
        <div className='h-full w-full flex'>
            <Routes>
                <Route path="/create" element={<CreateEnclaveModal addEnclave={addEnclave} name={name} setName={setName} args={args} setArgs={setArgs} handleSubmit={handleModalSubmit}/>}/>
                <Route path="/progress" element={<CreateEnclaveView args={args} packageId={name} enclave={enclave}/>} />
            </Routes>
        </div>
    )
}

export default CreateEnclave;
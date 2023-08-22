import React from 'react';
import {useNavigate} from 'react-router-dom';
import NoData from './NoData';
import WelcomePanel from "./Test";
import {useAppContext} from "../context/AppState";
import app from "../App";

const Main = ({totalEnclaves}) => {
    const {appData, setAppData} = useAppContext()
    const navigate = useNavigate()
    const handleCreateEnvClick = () => {
        navigate("/enclave/create")
    }
    const handleViewEnvsClick = () => {
        navigate("/enclaves")
    }

    const receiveMessage = (event) => {
        const message = event.data.message;
        switch (message) {
            case 'jwtToken':
                const value = event.data.value
                if (value !== null && value !== undefined) {
                    // console.log("Got the message!!", value)
                    setAppData({
                        ...appData,
                        jwtToken: value,
                    })
                }
                break;
        }
    }
    window.addEventListener("message", receiveMessage)

    return (
        <div className="flex-grow bg-#181926-100 flex-row flex mt-28 w-screen">
            {/*JWT: {appData.jwtToken}*/}
            <div className='w-1/3'></div>
            <div className="flex flex-col min-w-fit w-1/3">
                <div className='flex justify-center items-center w-full'>
                    <div className='text-center w-full'>
                        {
                            totalEnclaves > 0 ?
                                <div className="mb-4 bg-[#24BA27] h-16 rounded" onClick={handleViewEnvsClick}>
                                    <div className='cursor-default text-3xl text-slate-800 p-2'> View Enclave</div>
                                </div> : <NoData text={"No Enclaves Created"}/>
                        }
                        <div className="mb-4 bg-[#24BA27] h-16 rounded" onClick={handleCreateEnvClick}>
                            <div className='cursor-default text-3xl text-slate-800 p-2'> Create Enclave</div>
                        </div>
                    </div>
                </div>
            </div>
            <div className='w-1/3'></div>
        </div>
    );
}
export default Main;

import React from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

const Main = () => {
  
  const navigate = useNavigate()
  
  const handleCreateEnvClick = () => {
    navigate("/enclaves/create")
  }

  const handleViewEnvsClick = () => {
    navigate("/enclaves")
  }

  return (
    <div className="flex h-screen bg-slate-800 flex-row mt-28">
        <div className='w-1/3'></div>
        <div className="flex flex-col min-w-fit w-1/3">
            <div className='flex justify-center items-center w-full'>
                <div className='text-center w-full'>
                    <div className="mb-4 bg-green-800 h-16 rounded" onClick={handleViewEnvsClick}>
                        <div className='text-3xl p-2'> View Environments </div>
                    </div>
                    <div className="mb-4 bg-green-800 h-16 rounded" onClick={handleCreateEnvClick}>
                        <div className='text-3xl p-2'> Create Environment </div>
                    </div>
                </div>
            </div>
        </div>
        <div className='w-1/3'></div>
    </div>
  );
}

export default Main;

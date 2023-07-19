import React from 'react';
import { useNavigate } from 'react-router-dom';
import NoData from './NoData';

const Main = ({totalEnclaves}) => {
  const navigate = useNavigate()
  // const handleCreateEnvClick = () => {
  //   navigate("/enclaves/create")
  // }
  const handleViewEnvsClick = () => {
    navigate("/enclaves")
  }
  return (
    <div className="flex-grow bg-slate-800 flex-row flex mt-28 w-screen">
        <div className='w-1/3'> </div>
        <div className="flex flex-col min-w-fit w-1/3">
          <div className='flex justify-center items-center w-full'>
            <div className='text-center w-full'>
                {
                  totalEnclaves > 0 ? <div className="mb-4 bg-green-600 h-16 rounded" onClick={handleViewEnvsClick}>
                  <div className='cursor-default text-3xl text-slate-800 p-2'> View Environments </div>
              </div> : <NoData text={"No Enclaves Created"}/> 
                }
                {/* <div className="mb-4 bg-green-600 h-16 rounded" onClick={handleCreateEnvClick}>
                    <div className='cursor-default text-3xl text-slate-800 p-2'> Create Environment </div>
                </div> */}
            </div>
            </div>
        </div>
        <div className='w-1/3'></div>
    </div>
  );
}
export default Main;

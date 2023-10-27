import React, {useState}from 'react';
import { useNavigate } from 'react-router';
import {createEnclave} from "../api/enclave";
import {Button, Checkbox} from "@chakra-ui/react";

export const CreateEnclaveModal = ({enclaveName, handleSubmit, name, setName, args, setArgs, addEnclave, token, apiHost, setEnclaveName, productionMode, setProductionMode}) => {
  const [jsonError, setJsonError] = useState("")
  const navigate = useNavigate();
  const [runningPackage, setRunningPackage] = useState(false)

  const handleFormSubmit = (e) => {
    e.preventDefault();
    const fetch = async () => {
      try {
        const enclave = await createEnclave(token, apiHost, enclaveName, productionMode);
        addEnclave(enclave)
        handleSubmit(enclave);
      } catch(ex) {
        console.error(ex);
        alert(`Error occurred while creating enclave for package: ${name}. An error message should be printed in console, please share it with us to help debug this problem`)
      } finally {
        setRunningPackage(false)
      } 
    }

    try {
      setJsonError("")
      JSON.parse(args)
      setRunningPackage(true)
      fetch();
    } catch (error) {
      setJsonError("Invalid Json")
    }
  };

  return (
    <div className="flex justify-center w-full h-fit m-14">
      <form className="bg-gray-100 p-6 rounded-lg shadow-md w-1/3" on>
        <div className="text-center">
        <label className="block mb-4 text-2xl text-black">
            Enclave Name:
            <input
              type="text"
              placeholder="Optional"
              value={enclaveName}
              onChange={(e) => setEnclaveName(e.target.value)}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
        </label>
        <label className="block mb-4 text-2xl text-black space-x-2">
            <Checkbox colorScheme='blue' border={'black'} isChecked={productionMode} onChange={(e)=>setProductionMode(e.target.checked)}> 
                <div className="text-xl text-black"> Production Mode </div>
            </Checkbox>
        </label>
        <label className="block mb-4 text-2xl text-black">
            Package Id:
            <input
              type="text"
              placeholder="Required"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
        </label>
        <label className="block mb-4 text-black">
            Args: (Json)
            <textarea
              value={args}
              onChange={(e) => { setArgs(e.target.value); setJsonError("")}}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 resize-none focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              rows={4}
            ></textarea>
            <p className='text-red-300 font-bold'>{jsonError}</p>
        </label>
        <div className="flex row gap-10">
            <Button
              onClick={handleFormSubmit}
              type="submit"
              bg='blue.500'
              _hover={{ bg: "gray.500", svg: { fill: "black" } }}
              w="50%"
              isLoading={runningPackage}
              loadingText="Running..."
            >
              Run 
            </Button>
            <button
              type="back"
              className="w-[50%] bg-gray-500 text-white py-2 px-4 rounded-md hover:bg-gray-700 focus:outline-none focus:ring-blue-500 focus:ring-2 focus:ring-offset-2"
              onClick={() => navigate("/catalog", {replace:true})}
            >
              Package Catalog
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

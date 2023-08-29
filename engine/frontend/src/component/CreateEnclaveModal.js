import React, {useState}from 'react';
import { useNavigate } from 'react-router';
import {createEnclave} from "../api/enclave";

export const CreateEnclaveModal = ({handleSubmit, name, setName, args, setArgs, addEnclave, token, apiHost}) => {
  const [jsonError, setJsonError] = useState("")
  const navigate = useNavigate();

  const handleFormSubmit = (e) => {
    e.preventDefault();
    const fetch = async () => {
      const enclave = await createEnclave(token, apiHost);
      addEnclave(enclave)
      handleSubmit(enclave);
    }

    try {
      setJsonError("")
      JSON.parse(args)
      fetch();
    } catch (error) {
      setJsonError("Invalid Json")
    }
  };

  return (
    <div className="flex justify-center w-full h-fit m-14">
      <form className="bg-gray-100 p-6 rounded-lg shadow-md w-1/3" on>
        <div className="text-center">
          <label className="block mb-4 text-2xl">
            Package Id:
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </label>
          <label className="block mb-4">
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
            <button
              onClick={handleFormSubmit}
              type="submit"
              className="w-[50%] bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-blue-500 focus:ring-2 focus:ring-offset-2"
            >
              Submit
            </button>
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

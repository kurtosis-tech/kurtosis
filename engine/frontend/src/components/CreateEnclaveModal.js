import React, { useState } from 'react';
import {createEnclave} from "../api/enclave";

export const CreateEnclaveModal = ({handleSubmit, name, setName}) => {
  const handleFormSubmit = (e) => {
    e.preventDefault();
    const fetch = async () => {
      const enclaveInfo = await createEnclave();
      handleSubmit(enclaveInfo);
    }
    fetch();
  };

  return (
    <div className="flex justify-center items-center h-full">
      <form onSubmit={handleFormSubmit} className="bg-gray-100 p-6 rounded-lg shadow-md">
        <div className="text-center">
          <h2 className="text-2xl mb-4">{name}</h2>
          <label className="block mb-4">
            Name:
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </label>
          {/* <label className="block mb-4">
            Args (JSON format):
            <textarea
              value={args}
              onChange={(e) => setArgs(e.target.value)}
              className="block w-full rounded-md border-gray-300 py-2 px-3 mt-1 resize-none focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              rows={4}
            ></textarea>
          </label> */}
          <button
            type="submit"
            className="bg-blue-500 text-white py-2 px-4 rounded-md hover:bg-blue-600 focus:outline-none focus:ring-blue-500 focus:ring-2 focus:ring-offset-2"
          >
            Submit
          </button>
        </div>
      </form>
    </div>
  );
};
import React from 'react';

const LoadingOverlay = () => {
  return (
    <div className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center bg-black bg-opacity-50 z-50">
      <div className="border-4 border-gray-200 border-t-blue-500 rounded-full w-20 h-20 animate-spin"></div>
    </div>
  );
};

export default LoadingOverlay;

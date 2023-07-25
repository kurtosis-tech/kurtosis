import React from 'react';

const LoadingOverlay = () => {
  return (
    <div className="h-full w-full flex items-center justify-center">
      <div className="border-4 border-gray-200 border-t-blue-500 rounded-full w-20 h-20 animate-spin"></div>
    </div>
  );
};

export default LoadingOverlay;

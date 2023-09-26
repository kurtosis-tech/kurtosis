import React from 'react';

const NoData = ({text="No Data Available", color="text-slate-500", size="text-4xl"}) => {
  return (
    <div className="flex justify-center my-36">
      <p className={`${size} ${color} font-bold`}>{text}</p>
    </div>
  );
};

export default NoData;

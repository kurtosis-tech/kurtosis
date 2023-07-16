export const LeftView = ({heading, renderList}) => {
    // const renderServices = (services, handleClick) => {
    //     return services.map(service => {
    //         return (
    //             <div 
    //                 className={`cursor-default flex text-white items-center justify-center h-14 rounded-md border-4 bg-green-700`} 
    //                 key={service.uuid} onClick={()=>handleClick(service)}>
    //                 {service.name}
    //             </div>
    //         )   
    //     }) 
    // }

    return (
        <div className='flex flex-col bg-slate-800'>
            <div className="text-3xl text-center mt-5 mb-3 text-white">
                {heading} 
            </div>
            <div className="flex flex-col space-y-4 p-2 overflow-auto">
                {
                    renderList()                
                }
            </div> 
            {/* <div className="flex text-3xl justify-center py-10 text-white border-t-4"> {username} </div> */}
        </div>
    )
}
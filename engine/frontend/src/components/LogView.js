export const LogView = ({heading, logs, classAttr}) => {
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
        <div className={classAttr}>
            <div className="text-2xl text-center h-fit mb-2"> 
               {heading}
            </div>
            <div className="overflow-auto h-fit">
                {
                    logs.length === 0 ? 
                        <div className="font-bold text-3xl text-slate-400 text-center m-[10%] justify-center">
                            No Data
                        </div> : 
                    logs.map(
                        log => (
                            <div className="h-fit border-b-2 text-lg">
                                {log}
                            </div>
                        )
                    )
                }
            </div>
        </div> 
    )
}
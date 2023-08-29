import Heading from "./Heading"

export const LogView = ({heading, logs, size="h-[70%]" }) => {
    return (
        <div className={`flex-col flex ${size}`}>
            <Heading content={heading} />
            <div className="overflow-y-auto h-full">
                <ul className="border border-gray-200 p-2">
                    {logs.map((log, index) => (
                        <li key={index} className="p-2 border-b border-gray-200 text-black">
                        {log}
                        </li>
                    ))}
                </ul>
            </div>
        </div>    
    )
}   

{/* </div>
        <div className="h-full flex flex-col">
            <Heading content={heading} />
            <div className="overflow-auto flex-col h-4/6">
                {
                        logs.length === 0 ? 
                            <div className="font-bold text-3xl text-slate-400 text-center m-[10%] justify-center">
                                No Data
                            </div> : 
                        logs.map(
                            (log,index) => (
                                <div key={index} className="border-b-2 text-xl h-fit p-2">
                                    {log}
                                </div>
                            )
                        )
                }
            </div>
        </div> */}
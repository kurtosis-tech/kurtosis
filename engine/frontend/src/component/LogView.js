import Heading from "./Heading"

export const LogView = ({heading, logs, classAttr}) => {
    return (
        <div className={classAttr}>
            <Heading content={heading}/>
            <div className="overflow-y-scroll">
                {
                    logs.length === 0 ? 
                        <div className="font-bold text-3xl text-slate-400 text-center m-[10%] justify-center">
                            No Data
                        </div> : 
                    logs.map(
                        log => (
                            <div className="border-b-2 text-xl h-fit p-2">
                                {log}
                            </div>
                        )
                    )
                }
            </div>
        </div> 
    )
}
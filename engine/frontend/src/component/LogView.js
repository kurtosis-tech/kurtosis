import Heading from "../component/Heading"
import LoadingOverlay from "./LoadingOverflow"
import NoData from "./NoData"

const LogViewComponent = ({logs}) => (
    <div className="overflow-y-scroll">
        {
            logs.length === 0 ? 
                <NoData text={"No Logs Available"} size={"text-2xl"}/>
            : 
            logs.map(
                log => (
                    <div className="border-b-2 text-xl h-fit p-2">
                        {log}
                    </div>
                )
            )
        }
    </div>
)

export const LogView = ({heading, logs, classAttr, loading}) => {
    return (
        <div className={classAttr}>
            <Heading content={heading}/>
            {loading ? <LoadingOverlay /> : <LogViewComponent logs={logs}/>}
        </div> 
    )
}
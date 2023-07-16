import Heading  from "./Heading";

const LeftPanel = ({heading, renderList, home}) => {
    if (home) {
        return (
            <div>
            </div>
    )}

    return (
        <div className="flex-none bg-slate-800 w-1/6 border-r border-gray-300 min-w-fit">
            <Heading content={heading} color={"text-green-600"} />
            {
                (renderList) ?  
                <div className="h-full p-4 overflow-auto">
                    <div className="space-y-4">
                        {renderList()}
                    </div>
                </div>: 
                <div></div>
            }
           
        </div>
    )
}

export default LeftPanel;
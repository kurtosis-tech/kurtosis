import Heading  from "./Heading";

const LeftPanel = ({heading, renderList, home}) => {
    if (home) {
        return (
            <div>
            </div>
    )}

    return (
        <div className="flex-none bg-slate-800 w-[22rem] border-r border-gray-300">
            <Heading content={heading} color={"text-green-600"} />
            {
                (renderList) ?  
                <div className="h-5/6 m-4 p-2 overflow-auto">
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
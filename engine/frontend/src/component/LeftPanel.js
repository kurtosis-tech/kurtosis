import Heading  from "./Heading";

const LeftPanel = ({heading, renderList, home}) => {
    if (home) {
        return (
            <div>
            </div>
    )}

    return (
        <div className="flex-none bg-[#171923] w-[22rem]">
            <Heading content={heading} color={"text-[#24BA27]"} />
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

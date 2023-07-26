const Heading = ({content, size="text-2xl", padding="p-4", margin="m-2", color="text-black"}) => {
    const css = `${size} ${padding} ${margin} text-center font-bold ${color}`
    return (
        <div>
            <h2 className={`${css}`}> {content} </h2>
        </div>
    )
}

export default Heading;
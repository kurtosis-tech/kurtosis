import Heading from "./Heading";
import {useEffect, useState} from "react";
import {useLocation, useNavigate, useParams} from "react-router-dom";
import {getFileArtifactInfo} from "../api/container";

import NoData from "./NoData";
import LeftPanel from "./LeftPanel";
import RightPanel from "./RightPanel";
import LoadingOverlay from "./LoadingOverflow";
import {useAppContext} from "../context/AppState";
import {CodeEditor} from "./CodeEditor";
import {Box} from "@chakra-ui/react";

const BreadCrumbs = ({currentPath, handleOnClick, handleCleanButton}) => {
    const total = currentPath.length;

    const BreadCrumb = ({text, last, color = "text-white", index, handleOnClick}) => {
        return (
            <div className={`${color} cursor-default font-bold`}
                 onClick={() => handleOnClick(index)}
            >
                {text} {last ? "" : "/"}
            </div>)
    }

    return (
        <div className="flex flex-row py-2 px-5 flex-wrap">
            {
                currentPath.map((path, index) => (
                    <BreadCrumb key={index} text={path} last={total - 1 === index} index={index}
                                handleOnClick={handleOnClick}/>
                ))
            }
            {
                currentPath.length > 0 ?
                    <div className="mx-3" onClick={handleCleanButton}>
                        <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                             stroke="currentColor" aria-hidden="true">
                            <path color="gray" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                  d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </div> : null
            }
        </div>
    )
}

const renderFileArtfiacts = (fileArtifacts, handleClick) => {
    return fileArtifacts.map(fileArtifact => {
        return (
            <div className={`flex items-center justify-center h-14 text-base bg-[#24BA27]`} key={fileArtifact.name}
                 onClick={() => handleClick(fileArtifact.name)}>
                <div className='cursor-default text-lg text-white'> {fileArtifact.name} </div>
            </div>
        )
    })
}

const renderFiles = (files, handleFileClick) => {
    if (files.length === 0) {
        return (
            <NoData
                text={`No File Preview Available`}
                size={`text-xl`}
                color={`text-red-400`}
            />
        )
    }

    return Object.keys(files).map((key) => {
        return (
            <div className="border-2 bg-[#171923] text-lg align-middle text-center h-16 p-3 text-[#24BA27]"
                 onClick={() => handleFileClick(key, files[key])}
            >
                <div> {key} </div>
            </div>
        )
    })
}

const FileArtifactInfo = ({enclaves}) => {
    const navigate = useNavigate();

    const params = useParams();
    const {name: enclaveName, fileArtifactName} = params;
    const [fileInfoLoading, setFileInfoLoading] = useState(false);
    const [currentFiles, setCurrentFiles] = useState({})
    const [files, setFiles] = useState({})
    const [currentPath, setCurrentPath] = useState([])
    const [detailInfo, setDetailInfo] = useState({})
    const {state} = useLocation();
    const {fileArtifacts} = state;
    const {appData} = useAppContext()

    useEffect(() => {
        if (enclaves.length === 0) {
            navigate(`/enclaves/${enclaveName}`)
        } else {
            setFileInfoLoading(true)
            const fetch = async () => {
                const selected = enclaves.filter(enclave => enclave.name === enclaveName);
                if (selected.length > 0) {
                    const {files} = await getFileArtifactInfo(selected[0].host, selected[0].port, fileArtifactName, appData.jwtToken, appData.apiHost);
                    setFiles(files)
                    setCurrentFiles(files)
                    setCurrentPath([])
                    setDetailInfo({})
                }
                setFileInfoLoading(false)
            }
            fetch()
        }
    }, [fileArtifactName, appData.jwtToken])

    const handleCleanButton = () => {
        setCurrentPath([])
        setDetailInfo({})
        setCurrentFiles(files)
    }

    const getExtension = (filename) => {
        if (!filename || filename === undefined) return undefined
        return filename.split('.').pop();
    }

    const handleFileClick = (key, file) => {
        const extension = getExtension(file.path)
        file = {
            ...file,
            extension: extension,
        }
        if (file.path) {
            setDetailInfo(file)
            let current = files
            currentPath.map(path => {
                current = current[path]
            })
            setCurrentPath(c => [...c, key])
        } else {
            let current = files
            currentPath.map(path => {
                current = current[path]
            })
            setCurrentPath(c => [...c, key])
            setCurrentFiles(current[key])
            setDetailInfo({})
        }
    }

    const handleBreadCrumbClick = (index) => {
        if (index == currentPath.length - 1) {
            // do nothing
        } else {
            const newCurrentPath = currentPath.slice(0, index + 1)
            let current = files
            newCurrentPath.map(path => {
                current = current[path]
            })
            setCurrentFiles(current)
            setCurrentPath(newCurrentPath)
            setDetailInfo({})
        }
    }

    const handleLeftPanelClick = (fileArtifactName) => {
        navigate(`/enclaves/${enclaveName}/files/${fileArtifactName}`, {replace: true, state: {fileArtifacts}})
    }

    const FileInfoComponent = ({files, handleFileClick, detailInfo, fileName}) => {
        console.log(detailInfo)
        return (
            <div className='flex flex-col h-[90%] space-y-1 overflow-auto'>
                {
                    (Object.keys(detailInfo).length !== 0) ?
                        <div className="flex h-3/4 flex-col text-white">
                            <p className="text-sm font-bold text-left"> Size: {detailInfo.size?.toString()} B </p>
                            <p className="text-sm font-bold text-left"> Type: {detailInfo.extension} </p>
                            {detailInfo.textPreview.length > 0 ?
                                <Box>
                                    {
                                        CodeEditor(
                                            () => {
                                            },
                                            true,
                                            detailInfo.path,
                                            [detailInfo.extension],
                                            500,
                                            detailInfo.textPreview,
                                            false,
                                        )
                                    }
                                </Box>
                                :
                                <p className="break-all overflow-y-auto">
                                    <h2 className="text-2xl text-center mt-20 text-red-800 font-bold">
                                        No Preview Available
                                    </h2>
                                </p>
                            }
                        </div> :
                        <div className="flex flex-col h-[85%] min-h-[85%] border-2">
                            <Heading color="text-white" content={"Files"} size={"text-xl"}/>
                            <div className="overflow-auto space-y-2">
                                {renderFiles(files, handleFileClick)}
                            </div>
                        </div>
                }
            </div>
        )
    }

    return (
        <div className="flex h-full">
            <LeftPanel
                home={false}
                heading={"File Artifacts"}
                renderList={() => renderFileArtfiacts(fileArtifacts, handleLeftPanelClick)}
            />

            <div className="flex bg-[#171923] w-[calc(100vw-39rem)] flex-col space-y-5">
                <div className="h-[3rem] flex items-center justify-center m-2">
                    <Heading color="text-white" content={`${enclaveName} - ${fileArtifactName}`}/>
                </div>
                <BreadCrumbs
                    currentPath={currentPath}
                    handleOnClick={handleBreadCrumbClick}
                    handleCleanButton={handleCleanButton}
                />
                {fileInfoLoading ?
                    <LoadingOverlay/> :
                    <FileInfoComponent
                        files={currentFiles}
                        handleFileClick={handleFileClick}
                        detailInfo={detailInfo}
                        fileName={fileArtifactName}
                    />
                }
            </div>

            <RightPanel isServiceInfo={true} enclaveName={enclaveName}/>
        </div>
    )
}

export default FileArtifactInfo;

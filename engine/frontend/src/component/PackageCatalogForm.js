import {
    Box,
    Button,
    Center,
    Checkbox,
    Flex,
    Grid,
    GridItem, HStack,
    Input, InputGroup, InputLeftAddon, Spacer,
    Stack,
    Text,
    Textarea,
    Tooltip,
    useClipboard, useDimensions,

} from '@chakra-ui/react'
import PackageCatalogOption from "./PackageCatalogOption";
import {useLocation, useNavigate} from "react-router";
import React, {useEffect, useRef, useState} from 'react';
import startCase from 'lodash/startCase'
import {InfoOutlineIcon} from '@chakra-ui/icons'
import {ObjectInput} from 'react-object-input'
// import MonacoEditor from "react-monaco-editor/lib/editor";
import Editor, {DiffEditor, useMonaco, loader} from '@monaco-editor/react';
import useWindowDimensions from "../utils/windowheight";


const yaml = require("js-yaml")

const JsonEditor = (
    dataCallback,
    readOnly = false,
    fieldName = "json_field.json"
) => {
    const [value, setValue] = useState("{\n}")
    const jsonClipboard = useClipboard("");
    const monacoRef = useRef(null);
    const {width, height} = useWindowDimensions();
    const elementRef = useRef()
    const dimensions = useDimensions(elementRef)
    const editorHeight = 300

    useEffect(() => {
        handleEditorChange(value)
    }, [])

    // let ignoreEvent = false;
    // const updateHeight = () => {
    //     const contentHeight = Math.min(1000, editor.getContentHeight());
    //     container.style.width = `${width}px`;
    //     container.style.height = `${contentHeight}px`;
    //     try {
    //         ignoreEvent = true;
    //         editor.layout({ width, height: contentHeight });
    //     } finally {
    //         ignoreEvent = false;
    //     }
    // };

    useEffect(() => {
        jsonClipboard.setValue(value)
        if (monacoRef.current) {
            // console.log('monacoRef.current', monacoRef.current)
            if (monacoRef.current.editor) {
                // console.log('monacoRef.current.editor', monacoRef.current.editor)
                // console.log('monacoRef.current.editor', monacoRef.current.editor.getEditors())
                monacoRef.current.editor.getEditors()[0].updateOptions({automaticLayout:true})
                // monacoRef.current.editor.getEditors()[0].layout({width: 200, height:200})
                // console.log("width", width)
                // console.log("height", height)
                if(monacoRef.current?.editor){
                    // console.log("dims", monacoRef.current.editor.getEditors()[0].getLayoutInfo())
                }

                console.log(`box dims: ${dimensions?.borderBox?.width} x ${dimensions?.borderBox?.height}`, dimensions?.borderBox)


                // monacoRef.current.editor.getEditors()[0].layout({width: 600, height:editorHeight})
                // monacoRef.current.editor.layout();
                // monacoRef.current.editor.getEditors()[0].layout({width: "autp", height:200})

            }
            // monacoRef.current.layout({ width: 0, height: 0 })
        }
    }, [value])

    const saveTextAsFile = (text, fileName) => {
        const blob = new Blob([text], {type: "text/plain"});
        const downloadLink = document.createElement("a");
        downloadLink.download = fileName;
        downloadLink.innerHTML = "Download File";
        if (window.webkitURL) {
            // No need to add the download element to the DOM in Webkit.
            downloadLink.href = window.webkitURL.createObjectURL(blob);
        } else {
            downloadLink.href = window.URL.createObjectURL(blob);
            downloadLink.onclick = (event) => {
                if (event.target) {
                    document.body.removeChild(event.target);
                }
            };
            downloadLink.style.display = "none";
            document.body.appendChild(downloadLink);
        }

        downloadLink.click();

        if (window.webkitURL) {
            window.webkitURL.revokeObjectURL(downloadLink.href);
        } else {
            window.URL.revokeObjectURL(downloadLink.href);
        }
    };

    function handleEditorChange(value) {
        try {
            setValue(value)
            const parsedJson = JSON.parse(value)
            const jsonCleanedMinified = JSON.stringify(parsedJson)
            const jsonCleanedFormatted = JSON.stringify(parsedJson, null, 2)
            dataCallback(jsonCleanedMinified)
        } catch (error) {
            // swallow
        }
    }

    // function handleEditorWillMount(monaco) {
    //     // here is the monaco instance
    //     // do something before editor is mounted
    //     monaco.languages.typescript.javascriptDefaults.setEagerModelSync(true);
    // }

    function handleEditorDidMount(editor, monaco) {
        // here is another way to get monaco instance
        // you can also store it in `useRef` for further usage
        monacoRef.current = monaco;
        monacoRef.current.editor.getEditors()[0].updateOptions({ scrollBeyondLastLine: false });

    }

    function handleDownload() {
        saveTextAsFile(value, fieldName)
    }

    // TODO: We can use this to display error messages
    // function handleEditorValidation(markers) {
    //     // model markers
    //     // markers.forEach(marker => console.log('onValidate:', marker.message));
    // }
    return (
        <Box
            border="1px"
            borderColor='gray.700'
            borderRadius="7"
            margin={"1px"}
            padding={1}
            ref={elementRef}
        >
            <Editor
                margin={1}
                height="300px"
                // width="300px"
                defaultLanguage="json"
                value={value}
                theme={"vs-dark"}
                onMount={handleEditorDidMount}
                onChange={handleEditorChange}
                // onValidate={handleEditorValidation}
                options={{
                    automaticLayout: true,
                    selectOnLineNumbers: true,
                    languages: ['json'],
                    readOnly: readOnly,
                    minimap: {
                        enabled: false
                    }
                }}
            />
            <Button
                margin={1}
                onClick={jsonClipboard.onCopy}
            >
                {jsonClipboard.hasCopied ? "Copied!" : "Copy"}
            </Button>
            <Button
                margin={1}
                onClick={handleDownload}

            >
                Download
            </Button>
        </Box>
    )
}

const KeyValueTable = (dataCallBack) => {
    const [value, setValue] = useState({})
    const clipboard = useClipboard(value);

    useEffect(() => {
        dataCallBack(JSON.stringify(value))
        clipboard.setValue(JSON.stringify(value, null, 2))
    }, [value])

    const paste = async () => {
        const clipboard = await window.navigator.clipboard.readText()
        try {
            const json = JSON.parse(clipboard)
            setValue(json)
            console.log(value)
            console.log(json)
        } catch (e) {
            alert("Could not process the content in the clipboard. Please verify it's valid JSON")
        }
    }

    return (
        <Box
            border="1px"
            borderColor='gray.700'
            borderRadius="7"
            margin={"1px"}
            padding={1}
        >
            <ObjectInput
                obj={value}
                onChange={setValue}
                renderItem={(key, value, updateKey, updateValue, deleteProperty) => (
                    <Box
                        margin={1}
                    >
                        <HStack
                            spacing={1}
                            direction="row"
                        >
                            <InputGroup>
                                <InputLeftAddon children='Key'/>
                                <Input
                                    type="text"
                                    value={key}
                                    onChange={e => updateKey(e.target.value)}
                                    size="md"
                                    variant='filled'
                                />
                            </InputGroup>

                            <InputGroup>
                                <InputLeftAddon children='Value'/>
                                <Input
                                    type="text"
                                    value={value || ''} // value will be undefined for new rows
                                    onChange={e => updateValue(e.target.value)}
                                    size="md"
                                    variant='filled'
                                />
                            </InputGroup>
                            <Button
                                onClick={deleteProperty}
                            >
                                x
                            </Button>
                        </HStack>
                    </Box>
                )}
                renderAdd={addItem => <Button margin={1} onClick={addItem}>Add item</Button>}
                // renderEmpty={() => <p></p>}
            />
            <Button
                margin={1}
                onClick={clipboard.onCopy}
            >
                {clipboard.hasCopied ? "Copied!" : "Copy"}
            </Button>
            <Button
                margin={1}
                onClick={paste}
            >
                Paste
            </Button>
        </Box>
    )
}
const renderArgs = (args, handleChange, formData, errorData) => {
    return args.map((arg, index) => {
        let dataType = "";
        switch (arg.type) {
            case "INTEGER":
                dataType = "INTEGER"
                break;
            case "STRING":
                dataType = "STRING"
                break
            case "BOOL":
                dataType = "BOOL"
                break
            case "FLOAT":
                dataType = "FLOAT"
                break
            case "KEY_VALUE":
                dataType = "KEY_VALUE"
                break
            default:
                dataType = "JSON"
        }

        // no need to show plan arg as it's internal!
        if (arg.name === "plan") {
            return
        }

        return (
            <Flex color={"white"} key={`entry-${index}`}>
                <Flex mr="3" direction={"column"} w="15%">
                    <Text marginLeft={3}
                          align={"right"}
                          fontSize={"l"}
                    >
                        {startCase(arg.name)}
                        <Text
                            fontSize={"l"}
                            as="span"
                            color={"red"}
                        >
                            <Tooltip label="Required variable">
                                <span>{arg.isRequired ? " *" : null}</span>
                            </Tooltip>

                        </Text>
                        <Tooltip label="Some text coming here soon">
                            <InfoOutlineIcon marginLeft={2}/>
                        </Tooltip>

                    </Text>
                    <Text marginLeft={3} as='kbd' fontSize='xs' align={"right"}>{dataType.toLowerCase()}</Text>
                </Flex>
                <Flex flex="1" mr="3" direction={"column"}>
                    {errorData[index].length > 0 ?
                        <Text marginLeft={3} align={"left"} fontSize={"xs"}
                              color="red.500"> {errorData[index]} </Text> : null}
                    {renderSingleArg(arg.name, dataType, errorData, formData, index, handleChange)}
                </Flex>
            </Flex>
        )
    })
}

const renderSingleArg = (fieldName, type, errorData, formData, index, handleChange) => {
    switch (type) {
        case "INTEGER":
        case "STRING":
        case "BOOL":
        case "FLOAT":
            return (
                <Input
                    color='gray.300'
                    onChange={e => handleChange(e.target.value, index)}
                    value={formData[index]}
                    borderColor={errorData[index] ? "red.400" : null}
                />
            )

        case "JSON":
            // https://github.com/microsoft/monaco-editor/blob/main/webpack-plugin/README.md#options
            return (
                // <Textarea
                //     borderColor={errorData[index] ? "red.400" : null}
                //     minHeight={"200px"}
                //     onChange={e => handleChange(e.target.value, index)}
                //     value={formData[index]}
                // />
                // <MonacoEditor
                //     width="800"
                //     height="600"
                //     language="javascript"
                //     theme="vs-dark"
                //     value={formData[index]}
                //     options={options}
                //     // onChange={::this.onChange}
                //     // editorDidMount={::this.editorDidMount}
                // />
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >
                    {JsonEditor(
                        (data) => handleChange(data, index),
                        false,
                        fieldName,
                    )}
                </Box>

            )
        case "KEY_VALUE":
            return (
                <Box
                    border={errorData[index] ? "1px" : null}
                    borderColor={errorData[index] ? "red.400" : null}
                >

                    {KeyValueTable((data) => handleChange(data, index))}
                </Box>

            )

        default:
            return <p key={`data-${index}`}>Unsupported data type encountered</p>
    }
}


const checkValidUndefinedType = (data) => {
    try {
        const val = yaml.load(data)
    } catch (ex) {
        return false;
    }
    return true;
}

const checkValidJsonType = (data) => {
    console.log("data", data)
    if (data === undefined || data === "undefined" || data.length === 0) {
        return false
    }

    try {
        const val = JSON.parse(data)
        if (Object.keys(val).length > 0) {
            return true;
        }
        return false
    } catch (ex) {
        if (data.includes("\"") || data.includes("\'")) {
            return false
        }
        return true;
    }
}

const checkValidStringType = (data) => {
    if (data === "undefined" || data.length === 0) {
        return false
    }

    try {
        const val = JSON.parse(data)
        if (typeof val === "string") {
            return true
        }
        return false
    } catch (ex) {
        if (data.includes("\"") || data.includes("\'")) {
            return false
        }
        return true;
    }
}

const checkValidIntType = (data) => {
    const isNumeric = (value) => {
        return /^-?\d+$/.test(value);
    }
    if (data === "undefined") {
        return false
    }
    try {
        const trimmedData = data.trim()
        return isNumeric(trimmedData)
    } catch (ex) {
        return false
    }
}

const checkValidFloatType = (data) => {
    const isValidFloat = (value) => {
        return !isNaN(Number(value))
    }
    if (data === "undefined") {
        return false
    }

    const trimmedData = data.trim()
    if (trimmedData.length === 0) {
        return false
    }

    return isValidFloat(trimmedData)
}

const checkValidBooleanType = (data) => {
    const trimData = data.trim()
    return ["TRUE", "FALSE"].includes(trimData.toUpperCase())
}

const PackageCatalogForm = ({handleCreateNewEnclave}) => {
    const navigate = useNavigate()
    const location = useLocation()
    const {state} = location;
    const {kurtosisPackage} = state

    // TODO: REMOVE, FOR TESTING:
    kurtosisPackage.args.forEach(item => {
        if (item.name === "remote_chains") {
            item.type = "KEY_VALUE"
            item.data = {"ab": "cd"}
        } else if (item.type === undefined) {
            item.type = "JSON"
        }
    })

    const [runningPackage, setRunningPackage] = useState(false)
    const [enclaveName, setEnclaveName] = useState("")
    const [productionMode, setProductionMode] = useState(false)

    let initialFormData = {}
    kurtosisPackage.args.forEach(
        (arg, index) => {
            if (arg.name !== "plan") {
                initialFormData[index] = ""
            }
        }
    )
    const [formData, setFormData] = useState(initialFormData)

    let initialErrorData = {}
    kurtosisPackage.args.forEach((arg, index) => {
        if (arg.name !== "plan") {
            initialErrorData[index] = ""
        }
    })
    const [errorData, setErrorData] = useState(initialErrorData)

    const handleFormDataChange = (value, index) => {
        const newData = {
            ...formData,
            [index]: value,
        }
        setFormData(newData)

        if (errorData[index]) {
            const newErrorData = {
                ...errorData,
                [index]: false,
            }
            setErrorData(newErrorData)
        }
    }

    const handleCancelBtn = () => {
        navigate("/catalog")
    }

    const handleRunBtn = () => {
        let errorsFound = {}

        Object.keys(formData).filter(key => {
            let type = kurtosisPackage.args[key]["type"]
            const required = kurtosisPackage.args[key]["isRequired"]

            // if it's optional and empty it's fine
            if (!required && formData[key].length === 0) {
                return
            }

            let valid = true
            if (type === "STRING") {
                valid = checkValidStringType(formData[key])
            } else if (type === "INTEGER") {
                valid = checkValidIntType(formData[key])
            } else if (type === "BOOL") {
                valid = checkValidBooleanType(formData[key])
            } else if (type === "FLOAT") {
                valid = checkValidFloatType(formData[key])
            } else if (type === "JSON" || type === "KEY_VALUE") {
                valid = checkValidJsonType(formData[key])
            } else {
                valid = checkValidUndefinedType(formData[key])
            }

            let typeToPrint = type
            if (type === undefined) {
                typeToPrint = "JSON"
            }

            if (type === "BOOL") {
                typeToPrint = "BOOLEAN (TRUE/FALSE)"
            }

            if (!valid) {
                errorsFound[key] = `Incorrect type: expected ${typeToPrint}`;
            }
        })

        Object.keys(formData).filter(key => {
            let valid = true;
            const required = kurtosisPackage.args[key]["isRequired"]
            if (required) {
                if (formData[key].length === 0) {
                    valid = false;
                }
            }

            if (!valid) {
                errorsFound[key] = `This field is required and cannot be empty`;
            }
        })

        if (Object.keys(errorsFound).length === 0) {
            setRunningPackage(true)
            let args = {}
            Object.keys(formData).map(key => {
                const argName = kurtosisPackage.args[key].name
                const value = formData[key]

                let val;
                let trimmedValue;
                if (value.length > 0) {
                    if (kurtosisPackage.args[key]["type"] === "INTEGER") {
                        trimmedValue = value.trim()
                        val = parseInt(value)
                        args[argName] = val
                    } else if (kurtosisPackage.args[key]["type"] === "BOOL") {
                        trimmedValue = value.trim()
                        val = value.toUpperCase()
                        args[argName] = (val === "TRUE") ? true : false
                    } else if (kurtosisPackage.args[key]["type"] === "FLOAT") {
                        trimmedValue = value.trim()
                        val = parseFloat(value)
                        args[argName] = val
                    } else if (kurtosisPackage.args[key]["type"] === "STRING") {
                        args[argName] = value
                    } else {
                        val = JSON.parse(value)
                        args[argName] = val
                    }
                }
            })

            const stringifiedArgs = JSON.stringify(args)
            const runKurtosisPackageArgs = {
                packageId: kurtosisPackage.name,
                args: stringifiedArgs,
            }

            handleCreateNewEnclave(runKurtosisPackageArgs, enclaveName, productionMode)

        } else {
            const newErrorData = {
                ...errorData,
                ...errorsFound
            }
            setErrorData(newErrorData)
        }
    }

    return (
        <div className='w-screen'>
            <Grid
                templateAreas={`"option"
                                "packageId"
                                "main"
                                "configure"`}
                gridTemplateRows={'60px 60px 1fr 60px'}
                gridTemplateColumns={'1fr'}
                h='100%'
                w='100%'
                color='blackAlpha.700'
                fontWeight='bold'
                gap={2}
            >
                <GridItem area={'option'} pt="1">
                    <PackageCatalogOption catalog={true}/>
                </GridItem>
                <GridItem area={'packageId'} p="1">
                    <Flex direction={"column"} gap={"2"}>
                        <Center>
                            <Text color={"white"} fontSize={"2xl"}> {kurtosisPackage.name} </Text>
                        </Center>
                        <Checkbox
                            marginLeft={2}
                            color={"white"}
                            fontSize={"xl"}
                            isChecked={productionMode}
                            onChange={(e) => setProductionMode(e.target.checked)}
                        >
                            <Text>
                                Restart services
                                <Tooltip
                                    label="When enabled, Kurtosis will automatically restart any services that crash inside the enclave">
                                    <InfoOutlineIcon marginLeft={2}/>
                                </Tooltip>

                            </Text>
                        </Checkbox>
                    </Flex>
                </GridItem>
                <GridItem area={'main'} h="90%" overflowY={"scroll"} mt="10">
                    <Stack spacing={4}>
                        <Flex color={"white"}>
                            <Flex mr="3" direction={"column"} w="15%">
                                <Text
                                    marginLeft={3}
                                    align={"right"}
                                    fontSize={"l"}
                                >Enclave Name
                                    <Tooltip label="Leave empty to auto-generate an enclave name">
                                        <InfoOutlineIcon marginLeft={2}/>
                                    </Tooltip>
                                </Text>
                            </Flex>
                            <Flex flex="1" mr="3" direction={"column"}>
                                <Input
                                    color='gray.300'
                                    value={enclaveName}
                                    onChange={(e) => setEnclaveName(e.target.value)}
                                />
                            </Flex>
                        </Flex>
                        {renderArgs(kurtosisPackage.args, handleFormDataChange, formData, errorData)}
                    </Stack>
                </GridItem>
                <GridItem area={'configure'} m="10px">
                    <Flex gap={5}>
                        <Button colorScheme='red' w="50%" onClick={handleCancelBtn}> Cancel </Button>
                        <Button bg='#24BA27'
                                w="50%"
                                onClick={handleRunBtn}
                                isLoading={runningPackage}
                                loadingText="Running..."
                        >
                            Run
                        </Button>
                    </Flex>
                </GridItem>
            </Grid>
        </div>
    );
};
export default PackageCatalogForm;






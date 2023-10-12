import React, {useEffect, useState} from "react";
import {Box, Button, HStack, Input, InputGroup, InputLeftAddon, Tooltip, useClipboard} from "@chakra-ui/react";
import {ObjectInput} from "react-object-input";

const KeyValueTable = ({
                           dataCallback,
                           defaultState = {}
                       }) => {
    let parsed_json = {}
    if (defaultState !== undefined && defaultState !== "") {
        parsed_json = JSON.parse(defaultState)
    }

    const [value, setValue] = useState(parsed_json)
    const clipboard = useClipboard(value);

    useEffect(() => {
        const serialized = JSON.stringify(value)
        dataCallback(serialized)
        clipboard.setValue(JSON.stringify(value, null, 2))
    }, [value])

    const paste = async () => {
        const clipboard = await window.navigator.clipboard.readText()
        try {
            const json = JSON.parse(clipboard)
            setValue(json)
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
                                    value={value || ""} // value will be undefined for new rows
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
                renderAdd={addItem => <Button size={"sm"} margin={1} onClick={addItem}>Add item</Button>}
            />
            <Button
                margin={1}
                size={"sm"}
                onClick={clipboard.onCopy}
            >
                <Tooltip label="Copy as JSON">
                    {clipboard.hasCopied ? "Copied!" : "Copy"}
                </Tooltip>

            </Button>
            <Button
                margin={1}
                size={"sm"}
                onClick={paste}
            >
                <Tooltip label='Paste as a JSON key value map, e.g. `{ "key_1": "value", "key_2": 1 }` '>
                    Paste
                </Tooltip>
            </Button>
        </Box>
    )
}

export default KeyValueTable;

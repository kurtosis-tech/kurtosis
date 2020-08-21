data='{
    "jsonrpc": "2.0",
    "method": "KurtosisAPI.StartService",
    "params": {
        "imageName": "nginxdemos/hello",
        "usedPorts": [
            ":80"
        ],
        "startCommand": [],
        "dockerEnvironmentVars": {},
        "testVolumeMountFilepath": "/nothing-yet"
    },
    "id": 1
}'

curl -XPOST 'http://127.0.0.1:8080/jsonrpc' -H "content-type: application/json" --data "${data}"

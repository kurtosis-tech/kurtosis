package mocks

const (
	commonStatsResultStr = `
        "stats": {
            "summary": {
                "bytesProcessedPerSecond": 1243011,
                "linesProcessedPerSecond": 3249,
                "totalBytesProcessed": 3060,
                "totalLinesProcessed": 9,
                "execTime": 0.002461764,
                "queueTime": 0.000062976,
                "subqueries": 1,
                "totalEntriesReturned": 9
            },
            "querier": {
                "store": {
                    "totalChunksRef": 0,
                    "totalChunksDownloaded": 0,
                    "chunksDownloadTime": 0,
                    "chunk": {
                        "headChunkBytes": 0,
                        "headChunkLines": 0,
                        "decompressedBytes": 0,
                        "decompressedLines": 0,
                        "compressedBytes": 0,
                        "totalDuplicates": 0
                    }
                }
            },
            "ingester": {
                "totalReached": 1,
                "totalChunksMatched": 0,
                "totalBatches": 1,
                "totalLinesSent": 9,
                "store": {
                    "totalChunksRef": 2,
                    "totalChunksDownloaded": 2,
                    "chunksDownloadTime": 169136,
                    "chunk": {
                        "headChunkBytes": 0,
                        "headChunkLines": 0,
                        "decompressedBytes": 3060,
                        "decompressedLines": 9,
                        "compressedBytes": 737,
                        "totalDuplicates": 0
                    }
                }
            },
            "cache": {
                "chunk": {
                    "entriesFound": 0,
                    "entriesRequested": 0,
                    "entriesStored": 0,
                    "bytesReceived": 0,
                    "bytesSent": 0,
                    "requests": 0
                },
                "index": {
                    "entriesFound": 0,
                    "entriesRequested": 0,
                    "entriesStored": 0,
                    "bytesReceived": 0,
                    "bytesSent": 0,
                    "requests": 0
                },
                "result": {
                    "entriesFound": 0,
                    "entriesRequested": 0,
                    "entriesStored": 0,
                    "bytesReceived": 0,
                    "bytesSent": 0,
                    "requests": 0
                }
            }
        }`

	MockedResponseBodyWithSeveralValuesStr = `{
    "status": "success",
    "data": {
        "resultType": "streams",
        "result": [
            {
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-1"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ],
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the second log line.\"}"
                    ],
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the third log line.\"}"
                    ]
                ]
            },
            {
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-2"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ],
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the second log line.\"}"
                    ],
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the third log line.\"}"
                    ],
					[
                        "1664289367000000000",
                        "{\"log\":\"This is the fourth log line.\"}"
                    ]
                ]
            },
			{
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-3"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ],
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the second log line.\"}"
                    ]
                ]
            }
        ],
        `+ commonStatsResultStr +`
    }
}`

	MockedResponseBodyWithOneLineValuesStr = `{
    "status": "success",
    "data": {
        "resultType": "streams",
        "result": [
            {
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-1"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ]
                ]
            },
            {
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-2"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ]
                ]
            },
			{
                "stream": {
                    "comKurtosistechContainerType": "user-service",
                    "comKurtosistechGuid": "test-user-service-3"
                },
                "values": [
                    [
                        "1664289367000000000",
                        "{\"log\":\"This is the first log line.\"}"
                    ]
                ]
            }
        ],
        `+ commonStatsResultStr +`
    }
}`
)

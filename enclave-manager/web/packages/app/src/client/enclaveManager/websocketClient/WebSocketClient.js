
import { defaultQueryParamSerializer, defaultQuerySerializer, deepObjectPath, createFinalURL } from "openapi-fetch"

export default function createWSClient(clientOptions) {
    const {
        wss: baseFetch = WebSocket,
        querySerializer: globalQuerySerializer,
        ...baseOptions
    } = clientOptions ?? {};
    let baseUrl = baseOptions.baseUrl ?? "";
    if (baseUrl.endsWith("/")) {
        baseUrl = baseUrl.slice(0, -1); // remove trailing slash
    }

    async function coreFetch(url, callback, fetchOptions) {
        const {
            wss = baseFetch,
            headers,
            params = {},
            parseAs = "json",
            querySerializer = globalQuerySerializer ?? defaultQuerySerializer,
            ...init
        } = fetchOptions || {};

        // URL
        const finalURL = createFinalURL(url, {
            baseUrl,
            params,
            querySerializer,
        });

        var ws
        try {
            ws = new wss(finalURL);
        } catch (error) {
            return { error: {}, data: null }
        }

        ws.onmessage = (event) => {
            callback({ error: {}, data: JSON.parse(event.data) })
        };

        ws.onclose = () => {
            console.log('Disconnected from server');
        };

    }

    return {
        /** Call a GET endpoint */
        async GET(url, callback, init) {
            coreFetch(url, callback, { ...init, callback, method: "GET" });
            return;
        },
    };
}

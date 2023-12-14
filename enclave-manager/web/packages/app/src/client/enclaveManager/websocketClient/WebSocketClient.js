
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

    async function* websocketMessagesGenerator(url, fetchOptions) {
        const {
            wss = baseFetch,
            headers,
            params = {},
            parseAs = "json",
            querySerializer = globalQuerySerializer ?? defaultQuerySerializer,
            abortSignal,
            ...init
        } = fetchOptions || {};

        // URL
        const finalURL = createFinalURL(url, {
            baseUrl,
            params,
            querySerializer,
        });

        var socket
        try {
            socket = new wss(finalURL);
        } catch (error) {
            return { error: {}, data: null }
        }

        if (abortSignal) {
            if (abortSignal.aborted) {
                socket.close();  // already aborted, fail immediately
            }
            abortSignal.addEventListener('abort', () => socket.close());
        }

        // Wait for the WebSocket connection to be open
        await new Promise(resolve => {
            socket.addEventListener('open', resolve);
        });

        try {
            while (socket.readyState === WebSocket.OPEN) {
                // Wait for the next message
                const message = await new Promise(resolve => {
                    socket.addEventListener('message', event => resolve(event));
                });

                // Yield the received message
                yield { error: undefined, data: JSON.parse(message.data) };
            }
        } finally {
            // Close the WebSocket connection when the generator is done
            socket.close();
        }
    };

    return {
        /** Call a WS endpoint */
        WS: async function* (url, init) {
            for await (const val of websocketMessagesGenerator(url, { ...init, method: "GET" })) {
                yield val;
            }
        },
    };
}

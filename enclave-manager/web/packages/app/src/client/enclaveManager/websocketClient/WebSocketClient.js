import { createFinalURL, defaultQuerySerializer } from "openapi-fetch";

export default function createWSClient(clientOptions) {
  const { baseUrl = "" } = clientOptions ?? {};
  if (baseUrl.endsWith("/")) {
    baseUrl = baseUrl.slice(0, -1); // remove trailing slash
  }

  async function* websocketMessagesGenerator(url, fetchOptions) {
    const { params = {}, querySerializer = defaultQuerySerializer, abortSignal, ...init } = fetchOptions || {};

    // build full URL
    const finalURL = createFinalURL(url, {
      baseUrl,
      params,
      querySerializer,
    });

    var socket;
    try {
      socket = new WebSocket(finalURL);
    } catch (error) {
      return { error: {}, data: null };
    }

    // Wait for the WebSocket connection to be open
    await new Promise((resolve) => {
      socket.addEventListener("open", resolve);
    });

    if (abortSignal) {
      // already aborted, fail immediately
      if (abortSignal.aborted) {
        console.warn(`Websocket on ${finalURL} got aborted before using. Closing it.`);
        socket.close();
      }

      // close later if aborted
      abortSignal.addEventListener("abort", () => {
        console.warn(`Websocket on ${finalURL} has been asked to abort. Closing it.`);
        socket.close();
      });
    }

    try {
      while (socket.readyState === WebSocket.OPEN) {
        // Wait for the next message
        const message = await new Promise((resolve) => {
          socket.addEventListener("message", (event) => resolve(event));
        });

        // Yield the received message
        yield { error: undefined, data: JSON.parse(message.data) };
      }
    } finally {
      // Close the WebSocket connection when the generator is done
      socket.close();
    }
  }

  return {
    /** Call a WS endpoint */
    WS: async function* (url, init) {
      for await (const val of websocketMessagesGenerator(url, { ...init, method: "GET" })) {
        yield val;
      }
    },
  };
}

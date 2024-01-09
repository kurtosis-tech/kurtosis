import type {
  ErrorResponse,
  FilterKeys,
  MediaType,
  PathsWithMethod,
  ResponseObjectMap,
  SuccessResponse,
} from "openapi-typescript-helpers";

import type { ParamsOption, QuerySerializer } from "openapi-fetch";

import { createFinalURL, defaultQuerySerializer } from "openapi-fetch";

export interface ClientOptions {
  baseUrl?: string;
}

export type RequestOptions<T> = ParamsOption<T> & {
  querySerializer?: QuerySerializer<T>;
  abortSignal?: AbortSignal;
};

type ReturnType<Paths extends {}, P extends PathsWithMethod<Paths, "get">> = "get" extends infer T
  ? T extends "get"
    ? T extends keyof Paths[P]
      ? Paths[P][T]
      : unknown
    : never
  : never;

type ParamsType<Paths extends {}, P extends PathsWithMethod<Paths, "get">> = RequestOptions<
  FilterKeys<Paths[P], "get">
>;

export type MessageResponse<T> =
  | {
      data: FilterKeys<SuccessResponse<ResponseObjectMap<T>>, MediaType>;
      error?: never;
      message: MessageEvent<any>;
    }
  | {
      data?: never;
      error: FilterKeys<ErrorResponse<ResponseObjectMap<T>>, MediaType>;
      message: MessageEvent<any>;
    };

// This implementation is based on the http version of the lib `openapi-fetch`
// https://github.com/drwpow/openapi-typescript/blob/main/packages/openapi-fetch/src/index.d.ts
export function createWSClient<Paths extends {}>(
  clientOptions?: ClientOptions,
): {
  WS: <P extends PathsWithMethod<Paths, "get">>(
    url: P,
    ...init: ParamsType<Paths, P>[]
  ) => AsyncGenerator<MessageResponse<ReturnType<Paths, P>>>;
} {
  var baseUrl = clientOptions?.baseUrl ?? "";
  if (baseUrl.endsWith("/")) {
    baseUrl = baseUrl.slice(0, -1); // remove trailing slash
  }

  return {
    /** Call a WS endpoint */
    WS: async function* (url, fetchOptions) {
      const { params = {}, querySerializer = defaultQuerySerializer, abortSignal } = fetchOptions || {};

      // build full URL
      const finalURL = createFinalURL(url.toString(), {
        baseUrl,
        params,
        querySerializer,
      });

      var socket: WebSocket;
      var isWSPaused = false;
      var controller = new AbortController();

      // Create and wait for the WebSocket connection to be open
      async function wsConnect(): Promise<WebSocket> {
        return new Promise((resolve, reject) => {
          try {
            let newSocket = new WebSocket(finalURL);
            newSocket.addEventListener("open", () => resolve(newSocket));
          } catch (error) {
            reject(error);
          }
        });
      }

      // Start the connection for the first time
      socket = await wsConnect();

      // Handle client side request to abort (via abort signal)
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

      let reconnectHandler = async () => {
        if (!abortSignal?.aborted) {
          console.warn("Websocket connection unexpectedly closed, reconnecting");
          await wsConnect().then((ws) => {
            socket = ws;
          });
          controller.abort();
          controller = new AbortController();
          isWSPaused = false;
        }
      };

      // reconnect on unexpected close (i.e. abortSignal not aborted)
      socket.addEventListener("close", reconnectHandler);

      // // Pause and resume WS connection when
      // document.addEventListener("visibilitychange", async () => {
      //   if (document.hidden) {
      //     console.debug("Lost focus on web UI. Closing Websocket connection to save resources, resuming later");
      //     isWSPaused = true;
      //     socket.close();
      //   } else {
      //     console.debug("Resuming Websocket connection");
      //     reconnectHandler();
      //   }
      // });

      // Abortable message listener used to exit the message promise in case the connection
      // in broken and we replace with a new connection
      function waitForMsg(signal: AbortSignal): Promise<MessageEvent<any>> {
        if (signal.aborted) {
          return Promise.reject("Signal already aborted");
        }
        return new Promise((resolve, reject) => {
          socket.addEventListener("message", (event: MessageEvent<any>) => resolve(event));
          // Listen for abort event on signal
          signal.addEventListener("abort", () => {
            reject("Received signal to aborted");
          });
        });
      }

      // the async loop generator
      try {
        let message: MessageEvent<any> | undefined;
        // the isWSPaused keep the loop active even if the connection got closed. The loop
        // only exit if it's not paused and the connection is closed.
        while (isWSPaused || socket.readyState === WebSocket.OPEN) {
          // Wait for the next message and skip if gets an undefined message (i.e. aborted message await)
          message = await waitForMsg(controller.signal).catch((_) => undefined);
          if (message === undefined) {
            continue;
          }
          // Yield the received message
          yield { error: undefined, data: JSON.parse(message.data), message: message };
        }
      } catch (error) {
        console.error(`Received an unexpected message from the channel on ${finalURL}:`);
        console.error(error);
      } finally {
        // Final close. But let's remove the reconnection handler first
        socket.removeEventListener("close", reconnectHandler);
        socket.close();
      }
    },
  };
}

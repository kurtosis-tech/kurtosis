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
  WS: <P extends PathsWithMethod<Paths, "get">, T extends keyof Paths[P]>(
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
      const { params = {}, querySerializer = defaultQuerySerializer, abortSignal, ...init } = fetchOptions || {};

      // build full URL
      const finalURL = createFinalURL(url.toString(), {
        baseUrl,
        params,
        querySerializer,
      });

      var socket: WebSocket;
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
          const message: MessageEvent<any> = await new Promise((resolve) => {
            socket.addEventListener("message", (event: MessageEvent<any>) => resolve(event));
          });

          // Yield the received message
          yield { error: undefined, data: JSON.parse(message.data), message: message };
        }
      } catch (error) {
        console.error(`Received an unexpected message from the channel on ${finalURL}:`);
        console.error(error);
      } finally {
        // Close the WebSocket connection when the generator is done
        socket.close();
      }
    },
  };
}

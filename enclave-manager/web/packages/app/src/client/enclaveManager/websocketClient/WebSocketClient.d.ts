import type {
    ErrorResponse,
    SuccessResponse,
    FilterKeys,
    MediaType,
    PathsWithMethod,
    ResponseObjectMap,
    OperationRequestBodyContent,
    HasRequiredKeys,
} from "openapi-typescript-helpers";
import WebSocket from 'ws';

// Note: though "any" is considered bad practice in general, this library relies
// on "any" for type inference only it can give. Same goes for the "{}" type.
/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/ban-types */

/** Options for each client instance */
export interface ClientOptions extends Omit<RequestInit, "headers"> {
    /** set the common root URL for all API requests */
    baseUrl?: string;
    /** custom fetch (defaults to globalThis.fetch) */
    wss?: typeof WebSocket;
    /** global querySerializer */
    querySerializer?: QuerySerializer<unknown>;

}

export type HeadersOptions =
    | HeadersInit
    | Record<string, string | number | boolean | null | undefined>;

export type QuerySerializer<T> = (
    query: T extends { parameters: any }
        ? NonNullable<T["parameters"]["query"]>
        : Record<string, unknown>,
) => string;

export type ParseAs = "json" | "text" | "blob" | "arrayBuffer" | "stream";

export interface DefaultParamsOption {
    params?: {
        query?: Record<string, unknown>;
    };
}

export type ParamsOption<T> = T extends {
    parameters: any;
}
    ? HasRequiredKeys<T["parameters"]> extends never
    ? { params?: T["parameters"] }
    : { params: T["parameters"] }
    : DefaultParamsOption;

export type RequestBodyOption<T> = OperationRequestBodyContent<T> extends never
    ? { body?: never }
    : undefined extends OperationRequestBodyContent<T>
    ? { body?: OperationRequestBodyContent<T> }
    : { body: OperationRequestBodyContent<T> };

export type FetchOptions<T> = RequestOptions<T> & Omit<RequestInit, "body">;

export type Abortable = {
    abortSignal?: AbortSignal;
}

export type FetchResponse<T> =
    | {
        data: FilterKeys<SuccessResponse<ResponseObjectMap<T>>, MediaType>;
        error?: never;
        response: Response;
    }
    | {
        data?: never;
        error: FilterKeys<ErrorResponse<ResponseObjectMap<T>>, MediaType>;
        response: Response;
    };

export type RequestOptions<T> = ParamsOption<T> & Abortable &
    RequestBodyOption<T> & {
        querySerializer?: QuerySerializer<T>;
        parseAs?: ParseAs;
        wss?: ClientOptions["wss"];
    };

export default function createWSClient<Paths extends {}>(
    clientOptions?: ClientOptions,
): {
    /** Call a GET endpoint */
    GET<P extends PathsWithMethod<Paths, "get">>(
        url: P,
        ...init: HasRequiredKeys<
            FetchOptions<FilterKeys<Paths[P], "get">>
        > extends never
            ? [(FetchOptions<FilterKeys<Paths[P], "get">> | undefined)?]
            : [FetchOptions<FilterKeys<Paths[P], "get">>]
    ): AsyncGenerator<FetchResponse<"get" extends infer T ? T extends "get" ? T extends keyof Paths[P] ? Paths[P][T] : unknown : never : never>>;
};

export { };

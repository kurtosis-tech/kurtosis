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

import type {
    ParseAs,
    FetchResponse,
    ParamsOption,
    QuerySerializer
} from "openapi-fetch"

export interface ClientOptions extends Omit<RequestInit, "headers"> {
    /** set the common root URL for all API requests */
    baseUrl?: string;
    /** global querySerializer */
    querySerializer?: QuerySerializer<unknown>;
}

export type FetchOptions<T> = RequestOptions<T> & Omit<RequestInit, "body">;

export type RequestOptions<T> = ParamsOption<T> & {
    querySerializer?: QuerySerializer<T>;
    parseAs?: ParseAs;
    abortSignal?: AbortSignal;
};

export default function createWSClient<Paths extends {}>(
    clientOptions?: ClientOptions,
): {
    /** Call a WS endpoint */
    WS<P extends PathsWithMethod<Paths, "get">>(
        url: P,
        ...init: HasRequiredKeys<
            FetchOptions<FilterKeys<Paths[P], "get">>
        > extends never
            ? [(FetchOptions<FilterKeys<Paths[P], "get">> | undefined)?]
            : [FetchOptions<FilterKeys<Paths[P], "get">>]
    ): AsyncGenerator<FetchResponse<"get" extends infer T ? T extends "get" ? T extends keyof Paths[P] ? Paths[P][T] : unknown : never : never>>;
};

export { };

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
    FetchResponse,
    ParamsOption,
    QuerySerializer
} from "openapi-fetch"

export interface ClientOptions {
    baseUrl?: string;
}

export type RequestOptions<T> = ParamsOption<T> & {
    querySerializer?: QuerySerializer<T>;
    abortSignal?: AbortSignal;
};

// This implementation is based on the http version of the lib `openapi-fetch`
// https://github.com/drwpow/openapi-typescript/blob/main/packages/openapi-fetch/src/index.d.ts
export default function createWSClient<Paths extends {}>(
    clientOptions?: ClientOptions,
): {
    /** Call a WS endpoint */
    WS<P extends PathsWithMethod<Paths, "get">>(
        url: P,
        ...init: HasRequiredKeys<
            RequestOptions<FilterKeys<Paths[P], "get">>
        > extends never
            ? [(RequestOptions<FilterKeys<Paths[P], "get">> | undefined)?]
            : [RequestOptions<FilterKeys<Paths[P], "get">>]
    ): AsyncGenerator<FetchResponse<"get" extends infer T ? T extends "get" ? T extends keyof Paths[P] ? Paths[P][T] : unknown : never : never>>;
};

export { };

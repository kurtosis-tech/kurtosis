const AUTH_TOKEN_ENV_VAR_NAME: string = "KURTOSIS_API_AUTH_TOKEN";
const APPLICATION_DIRNAME: string = "kurtosis";
const TOKEN_FILE_NAME: string = "api-auth-token";
const GRPC_METADATA_KEY: string = "authorization";
const BEARER_PREFIX: string = "Bearer ";

async function getConfigHomeDirpath(): Promise<string> {
    const os = await import( /* webpackIgnore: true */ "os");
    const path = await import( /* webpackIgnore: true */ "path");
    if (process.env.XDG_CONFIG_HOME) {
        return process.env.XDG_CONFIG_HOME;
    }
    if (process.platform === "darwin") {
        return path.join(os.homedir(), "Library", "Application Support");
    }
    if (process.platform === "win32") {
        return process.env.LOCALAPPDATA ?? path.join(os.homedir(), "AppData", "Local");
    }
    return path.join(os.homedir(), ".config");
}

export async function getAuthToken(): Promise<string> {
    const tokenFromEnv = (process.env[AUTH_TOKEN_ENV_VAR_NAME] ?? "").trim();
    if (tokenFromEnv !== "") {
        return tokenFromEnv;
    }
    const fs = await import( /* webpackIgnore: true */ "fs");
    const path = await import( /* webpackIgnore: true */ "path");
    const tokenFilepath = path.join(await getConfigHomeDirpath(), APPLICATION_DIRNAME, TOKEN_FILE_NAME);
    try {
        return fs.readFileSync(tokenFilepath, "utf8").trim();
    } catch (error) {
        return "";
    }
}

export async function getGrpcNodeAuthClientOptions(grpcNode: typeof import("@grpc/grpc-js")): Promise<{ interceptors: import("@grpc/grpc-js").Interceptor[] }> {
    const token = await getAuthToken();
    if (token === "") {
        return { interceptors: [] };
    }
    const authInterceptor: import("@grpc/grpc-js").Interceptor = (options, nextCall) => {
        return new grpcNode.InterceptingCall(nextCall(options), {
            start: (metadata, listener, next) => {
                metadata.set(GRPC_METADATA_KEY, BEARER_PREFIX + token);
                next(metadata, listener);
            },
        });
    };
    return { interceptors: [authInterceptor] };
}

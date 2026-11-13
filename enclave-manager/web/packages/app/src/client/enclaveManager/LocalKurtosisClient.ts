import { createPromiseClient, Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { KURTOSIS_EM_API_DEFAULT_URL } from "../constants";
import { KurtosisClient } from "./KurtosisClient";

const AUTH_TOKEN_FRAGMENT_PREFIX = "#kurtosis-token=";
const AUTH_TOKEN_STORAGE_KEY = "kurtosis-api-auth-token";

function loadAuthToken(): string | null {
  try {
    if (window.location.hash.startsWith(AUTH_TOKEN_FRAGMENT_PREFIX)) {
      const token = decodeURIComponent(window.location.hash.slice(AUTH_TOKEN_FRAGMENT_PREFIX.length));
      window.localStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token);
      window.history.replaceState(null, "", window.location.pathname + window.location.search);
    }
    return window.localStorage.getItem(AUTH_TOKEN_STORAGE_KEY);
  } catch (error) {
    return null;
  }
}

const authToken = loadAuthToken();

const authInterceptor: Interceptor = (next) => async (req) => {
  if (authToken) {
    req.header.set("Authorization", `Bearer ${authToken}`);
  }
  return next(req);
};

export class LocalKurtosisClient extends KurtosisClient {
  constructor() {
    const defaultUrl = new URL(`${window.location.protocol}//${window.location.host}`);
    var baseUrl = KURTOSIS_EM_API_DEFAULT_URL;
    if (window.env !== undefined && window.env.domain !== undefined) {
      baseUrl = "https://" + window.env.domain;
    }
    super(
      createPromiseClient(
        KurtosisEnclaveManagerServer,
        createConnectTransport({ baseUrl: baseUrl, interceptors: [authInterceptor] }),
      ),
      defaultUrl,
      defaultUrl,
    );
  }

  getHeaderOptions() {
    return {};
  }

  isRunningInCloud(): boolean {
    return false;
  }
}

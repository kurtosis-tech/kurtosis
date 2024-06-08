import { isDefined } from "kurtosis-ui-components";

// Configurable:
export const KURTOSIS_CLOUD_PROTOCOL = "https";
export const KURTOSIS_CLOUD_HOST = "cloud.kurtosis.com";
export const KURTOSIS_CLOUD_CONNECT_PAGE = "connect";
export const KURTOSIS_CLOUD_EM_PAGE = "enclave-manager";
export const KURTOSIS_CLOUD_SUBSCRIPTION_PAGE = "subscription";
export const KURTOSIS_CLOUD_INSTANCE_MAX_UPTIME_IN_HOURS = 12;

// Cloud
export const KURTOSIS_CLOUD_UI_URL =
  process.env.REACT_APP_KURTOSIS_CLOUD_UI_URL || `${KURTOSIS_CLOUD_PROTOCOL}://${KURTOSIS_CLOUD_HOST}`;
export const KURTOSIS_CLOUD_CONNECT_URL = `${KURTOSIS_CLOUD_UI_URL}/${KURTOSIS_CLOUD_CONNECT_PAGE}`;
export const KURTOSIS_CLOUD_SUBSCRIPTION_URL = `${KURTOSIS_CLOUD_UI_URL}/${KURTOSIS_CLOUD_SUBSCRIPTION_PAGE}`;
export const KURTOSIS_CLOUD_EM_URL = `${KURTOSIS_CLOUD_UI_URL}/${KURTOSIS_CLOUD_EM_PAGE}`;
export const KURTOSIS_PACKAGE_INDEXER_URL =
  process.env.REACT_APP_KURTOSIS_PACKAGE_INDEXER_URL || `${KURTOSIS_CLOUD_PROTOCOL}://${KURTOSIS_CLOUD_HOST}:9770`;

// EM API
export const KURTOSIS_EM_DEFAULT_HOST = process.env.REACT_APP_KURTOSIS_DEFAULT_HOST || "localhost";
export const KURTOSIS_DEFAULT_EM_API_PORT = isDefined(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  ? parseInt(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  : 8081;
export const KURTOSIS_EM_API_DEFAULT_URL =
  process.env.REACT_APP_KURTOSIS_DEFAULT_URL || `http://${KURTOSIS_EM_DEFAULT_HOST}:${KURTOSIS_DEFAULT_EM_API_PORT}`;

declare global {
  interface Window {
    env: {
      domain: string;
    };
  }
}

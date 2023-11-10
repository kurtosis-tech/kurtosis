import { isDefined } from "../utils";

// Configurable:
export const KURTOSIS_CLOUD_PROTOCOL = "https";
export const KURTOSIS_CLOUD_HOST = "cloud.kurtosis.com";
export const KURTOSIS_CLOUD_CONNECT_PAGE = "connect";

// Cloud
export const KURTOSIS_CLOUD_UI_URL = process.env.REACT_APP_KURTOSIS_CLOUD_UI_URL || `https://${KURTOSIS_CLOUD_HOST}`;
export const KURTOSIS_CLOUD_CONNECT_URL = `${KURTOSIS_CLOUD_PROTOCOL}://${KURTOSIS_CLOUD_HOST}/${KURTOSIS_CLOUD_CONNECT_PAGE}`;
export const KURTOSIS_PACKAGE_INDEXER_URL =
  process.env.REACT_APP_KURTOSIS_PACKAGE_INDEXER_URL || `https://${KURTOSIS_CLOUD_HOST}:9770`;

// EM API
export const KURTOSIS_EM_DEFAULT_HOST = process.env.REACT_APP_KURTOSIS_DEFAULT_HOST || "localhost";
export const KURTOSIS_DEFAULT_EM_API_PORT = isDefined(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  ? parseInt(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  : 8081;
export const KURTOSIS_EM_API_DEFAULT_URL = process.env.REACT_APP_KURTOSIS_DEFAULT_URL ||
  `http://${KURTOSIS_EM_DEFAULT_HOST}:${KURTOSIS_DEFAULT_EM_API_PORT}`;

import { isDefined } from "../utils";

export const KURTOSIS_EM_DEFAULT_HOST = process.env.REACT_APP_KURTOSIS_DEFAULT_HOST || "localhost";
export const KURTOSIS_DEFAULT_EM_API_PORT = isDefined(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  ? parseInt(process.env.REACT_APP_KURTOSIS_DEFAULT_EM_API_PORT)
  : 8081;
export const KURTOSIS_EM_API_DEFAULT_URL =
  process.env.REACT_APP_KURTOSIS_DEFAULT_URL || `http://${KURTOSIS_EM_DEFAULT_HOST}:${KURTOSIS_DEFAULT_EM_API_PORT}`;

export const KURTOSIS_PACKAGE_INDEXER_URL = process.env.REACT_APP_KURTOSIS_PACKAGE_INDEXER_URL || "https://cloud.kurtosis.com:9770";
export const KURTOSIS_CLOUD_HOST = process.env.REACT_APP_KURTOSIS_CLOUD_UI_URL || "https://cloud.kurtosis.com";

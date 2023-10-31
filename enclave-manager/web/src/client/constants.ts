import { isDefined } from "../utils";

export const KURTOSIS_DEFAULT_HOST = process.env.REACT_APP_KURTOSIS_DEFAULT_HOST || "localhost";
export const KURTOSIS_DEFAULT_PORT = isDefined(process.env.REACT_APP_KURTOSIS_DEFAULT_PORT)
  ? parseInt(process.env.REACT_APP_KURTOSIS_DEFAULT_PORT)
  : 8081;
export const KURTOSIS_DEFAULT_URL =
  process.env.REACT_APP_KURTOSIS_DEFAULT_URL || `http://${KURTOSIS_DEFAULT_HOST}:${KURTOSIS_DEFAULT_PORT}`;

export const KURTOSIS_CLOUD_URL = process.env.REACT_APP_KURTOSIS_CLOUD_URL || "https://cloud.kurtosis.com:9770";

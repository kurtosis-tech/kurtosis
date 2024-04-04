import Cookies from "js-cookie";

export const instanceUUID = Cookies.get("_kurtosis_instance_id") || "";
export const apiKey = Cookies.get("_kurtosis_api_key") || "";
export const jwtToken = Cookies.get("_kurtosis_jwt_token");
export const isPrevEnv = true;
//export const isPrevEnv = Cookies.get("_prevenv_web_key") !== undefined;

import Cookies from "js-cookie";

export const instanceUUID = Cookies.get("_kurtosis_instance_id") || "";
export const jwtToken = Cookies.get("_kurtosis_jwt_token");

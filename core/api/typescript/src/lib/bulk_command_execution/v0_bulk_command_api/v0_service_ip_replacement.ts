const SERVICE_ID_IP_REPLACEMENT_PREFIX: string = "<<<";
const SERVICE_ID_REPLACEMENT_SUFFIX: string = ">>>";

// Used to encode a service ID to a string that can be embedded in commands, and which the API container will replace
// with the IP address of the service at runtime
function encodeServiceIdForIpReplacement(serviceId: string): string {
	return SERVICE_ID_IP_REPLACEMENT_PREFIX + serviceId + SERVICE_ID_REPLACEMENT_SUFFIX;
}

package enclave

// We restrict enclave names to 60 characters bc K8s doesn't allow names beyond 63 chars so 63 - 3 ("kt-" prefix) = 60
// The enclave name has to be a valid DNS-1035 value ("." and "_"  are not allowed) because it's mandatory for creating K8s resources like the namespace name
const AllowedEnclaveNameCharsRegexStr = `^[-A-Za-z0-9]{1,60}$`

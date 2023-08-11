package enclave

// We restrict enclave names to 60 characters bc K8s doesn't allow names beyond 63 chars so 63 - 3 ("kt-" prefix) = 60
const AllowedEnclaveNameCharsRegexStr = `^[-A-Za-z0-9._]{1,60}$`

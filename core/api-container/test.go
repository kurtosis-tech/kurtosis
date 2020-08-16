package api_container

import "net/http"

type KevinTest struct {}

type InputArg struct {
	saying string
}

type OutputArg struct {
	response string
}

func (obj KevinTest) Parrot(httpReq *http.Request, args *InputArg, result *OutputArg) error {
	result.response = args.saying
	return nil
}

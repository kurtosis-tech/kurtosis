package main

import (
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"log"
	"net/http"
)

type TestHandler struct {}

type InputArg struct {
	Saying string		`json:"saying"`
}

type OutputArg struct {
	Response string 	`json:"response"`
}

func (obj *TestHandler) Parrot(httpReq *http.Request, args *InputArg, result *OutputArg) error {
	result.Response = args.Saying
	log.Printf("Result: %v", result)
	return nil
}

func main() {
	server := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterService(new(TestHandler), "")

	http.Handle("/jsonrpc", server)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
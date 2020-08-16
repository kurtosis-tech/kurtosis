package api_container

import (
	"github.com/gorilla/rpc/v2/json2"
	"log"
	"net/http"
	"github.com/gorilla/rpc"
)

func main() {
	server := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterService(KevinTest{}, "")

	http.Handle("/jsonrpc/", server)
	log.Fatal(http.ListenAndServe(":80", nil))
}
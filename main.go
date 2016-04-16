package main

import (
	"flag"
	"net/http"
)

// APIVer is the current version of Obozrenie server API.
const APIVer = "0.1"
const APIPrefix = "/" + APIVer

const gameCollPrefix = APIPrefix + "/gamecoll"
const systemPrefix = APIPrefix + "/system"

func main() {
	var sAddr = flag.String("addr", ":16987", "Server address")
	var authPass = flag.String("password", "", "Server access password")

	flag.Parse()

	var exitChan = make(chan struct{})

	var actions = makeActionInstance(*authPass)
	var sMux = makeServeMux(makeActionInstance(*authPass), exitChan)

	var server = &http.Server{
		Addr:    *sAddr,
		Handler: sMux,
	}
	go server.ListenAndServe()

	<-exitChan
	actions.cleanup()
}

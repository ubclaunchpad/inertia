package main

import (
	"log"
	"net/http"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is tls over tcp"))
}

func main() {
	http.HandleFunc("/", HelloServer)
	log.Fatal(http.ListenAndServeTLS(
		":443",
		"./server.cert",
		"./server.key",
		nil))
}

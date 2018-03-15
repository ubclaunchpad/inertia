package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is tls over tcp"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloServer)

	manager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: manager.GetCertificate,
		},
	}

	go http.ListenAndServe(":80", manager.HTTPHandler(nil))
	log.Fatal(server.ListenAndServeTLS("", ""))
}

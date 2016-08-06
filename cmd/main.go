package main

import (
	"log"
	"syscall"
	"os"
	"github.com/gearmover/web2tor/proxy"
)

func main() {
	log.Println("Web2Tor HTTP Proxy")

	domain := ".chr1s.co"

	if len(os.Args) > 1 {
		domain = "." + os.Args[1]
	}

	httpServer, err := proxy.NewHTTPServer("0.0.0.0:80", "127.0.0.1:9050")
	if err != nil {
		log.Println("[!] Error Creating HTTP Server : ", err.Error())
		return
	}

	httpsServer, err := proxy.NewHTTPSServer("0.0.0.0:443", "127.0.0.1:9050")
	if err != nil {
		log.Println("[!] Error Creating HTTPS Server : ", err.Error())
		return
	}

	syscall.Setuid(1000)
	syscall.Setgid(1000)

	go httpServer.ListenAndServe(domain)

	httpsServer.ListenAndServe(domain)
}

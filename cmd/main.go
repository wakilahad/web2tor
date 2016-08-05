package main

import (
	"fmt"
	"github.com/gearmover/web2tor/proxy"
)

func main() {
	fmt.Println("Web2Tor HTTP Proxy")

	httpServer := proxy.NewHTTPServer("0.0.0.0:80", "127.0.0.1:9050")
	httpsServer := proxy.NewHTTPSServer("0.0.0.0:443", "127.0.0.1:9050")
}

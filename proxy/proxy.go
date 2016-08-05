package proxy

import (
	"golang.org/x/net/proxy"
	"errors"
)

type Server struct {
	fromConn   net.Listener
	toConn 	   net.Conn
}

var (
	ErrGeneric = errors.New("Generic Network Error")
)

func NewHTTPServer(listenAddr, dialAddr string) (*Server, error) {

	dialer, err := proxy.SOCKS5("tcp", dialAddr, nil, nil)
	if err != nil {
		return nil, err
	}

	return nil, ErrGeneric
}

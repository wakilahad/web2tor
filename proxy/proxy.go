package proxy

import (
	"bytes"
	"io/ioutil"
	"golang.org/x/net/proxy"
	"regexp"
	"bufio"
	"net"
	"errors"
	"log"
	"net/http"
	"crypto/tls"
	"html/template"
)

type Server struct {
	fromConn   net.Listener
	dialConn   proxy.Dialer
}

var (
	ErrGeneric = errors.New("Generic Network Error")
)

func NewHTTPServer(listenAddr, dialAddr string) (*Server, error) {

	dialer, err := proxy.SOCKS5("tcp", dialAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		dialConn: dialer,
		fromConn: listener,
	}

	return s, nil
}

func NewHTTPSServer(listenAddr, dialAddr string) (*Server, error) {

	cer, err := tls.LoadX509KeyPair("server.pem", "server.key")
    if err != nil {
        return nil, err
    }

    config := &tls.Config{Certificates: []tls.Certificate{cer}}
    ln, err := tls.Listen("tcp", listenAddr, config)
    if err != nil {
        return nil, err
    }

	dialer, err := proxy.SOCKS5("tcp", dialAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	s := &Server{
		dialConn: dialer,
		fromConn: ln,
	}

	return s, nil
}

func (s *Server) ListenAndServe(domain string) error {

	defer s.fromConn.Close()

	if s == nil {
		return ErrGeneric
	}

	for {
		conn, err := s.fromConn.Accept()
		if err != nil {
			log.Println("[!] Unable to Accept Client : ", err.Error())
			continue
		}

		go s.HandleConn(conn, domain)
	}
}

func removeDuplicates(array [][]byte) [][]byte {
	uniq := make(map[string][]byte)

	for i:=0; i < len(array); i++ {
		uniq[string(array[i])] = array[i]
	}

	newArray := make([][]byte, 0, len(uniq))

	for k,_ := range uniq {
		newArray = append(newArray, uniq[k])
	}

	return newArray
}

func (s *Server) HandleConn(conn net.Conn, domain string) {

	// make sure to clean up the incoming connection
	defer conn.Close()

	// create source buffered IO reader/writer
	sReader := bufio.NewReader(conn)
	sWriter := bufio.NewWriter(conn)

	// read the HTTP request
	req, err := http.ReadRequest(sReader)
	if err != nil {
		log.Println("[!] Unable to Initialize Proxy Conn : ", err.Error())
		return
	}

	// DEBUG:
	//log.Printf("Request: %v\n", req)

	// seperates out the host and URI
	httpRE,err := regexp.Compile("([a-zA-Z0-9]+.onion){1}"+domain)
	if err != nil {
		log.Println("[!] Error Compiling URL Regular Expression : ", err.Error())
		return
	}

	// Regular Expression the URL
	matches := httpRE.FindSubmatch([]byte(req.Host))
	if err != nil {
		log.Println("[!] URL RE Failed to Match : ", err.Error())
		return
	}

	if len(matches) != 2 {
		log.Printf("[*] URL Matches : %v Host String : %s\n", matches, req.URL.Host)

		// load the 404 template
		tmpl, err := template.ParseFiles("../templates/404.html")
		if err != nil {
			sWriter.Write([]byte("Double 404 !  Pages Not Found !"))
			sWriter.Flush()
			return
		}

		tmpl.Execute(sWriter, struct{}{})
		sWriter.Flush()

		return
	}

	// set the host and uri variables
	Host := string(matches[1])
	URI  := req.URL.RequestURI()

	// create the SOCKS5 compliant URL
	URL := Host + ":80"

	// DEBUG:
	//log.Println("[*] Proxying URL : ", URL)
	//log.Println("[*] Req Cookies : ", req.Header)

	// create the SOCKS5 connection
	toConn, err := s.dialConn.Dial("tcp", URL)
	if err != nil {
		log.Println("[!] Unable to Dial TOR : ", err.Error())

		// load the 404 template
		tmpl, err := template.ParseFiles("../templates/404.html")
		if err != nil {
			sWriter.Write([]byte("Double 404 !  Pages Not Found !"))
			sWriter.Flush()
			return
		}

		tmpl.Execute(sWriter, struct{}{})
		sWriter.Flush()

		return
	}

	// reset the request Host to the original URL
	req.Host = Host

	// DEBUG: print out the request
	//log.Println("Request:", req)

	// clean up the outgoing connection
	defer toConn.Close()

	// create destination buffered IO
	dReader := bufio.NewReader(toConn)
	dWriter := bufio.NewWriter(toConn)

	// create the modified URL without the server host IP
	req.URL, err = req.URL.Parse("http://" + Host + URI)
	if err != nil {
		log.Println("[!] Unable to Reassign URL : ", err.Error())
		return
	}

	// write the request to the SOCKS5 proxy
	req.Write(dWriter)
	dWriter.Flush()

	// read the response from the SOCKS5 proxy
	resp, err := http.ReadResponse(dReader, req)
	if err != nil {
		log.Println("[!] Unable to Read Response : ", err.Error())
		return
	}

	// DEBUG:
	//log.Println("Response Code : ", resp.Status)

	if len(resp.Header.Get("Location")) > 0 {
		location := bytes.Replace([]byte(resp.Header.Get("Location")), []byte(Host), []byte(Host + domain), -1)
		resp.Header.Set("Location", string(location))
	}

	// parse the response and replace any anchor links with an updated URL
	buffer, _ := ioutil.ReadAll(resp.Body)

	// replace all host URL's with modified ones
	fromStrings := []string{
		Host,
	}

	toStrings := []string{
		Host + domain,
	}

	if len(fromStrings) != len(toStrings) {
		// PROGRAMMER ERROR: fix your from strings and to strings!
		panic("fix your from strings and to strings!")
	}

	// replace all the strings that need replacing in the returned document
	//log.Printf("Original Length: %d\n", len(tmpNewReader.Bytes()))
	//buffer = bytes.Replace(tmpNewReader.Bytes(), []byte(fromStrings[0]), []byte(toStrings[0]), -1)
	onionRE := regexp.MustCompile("[a-zA-Z0-9]+.onion")
	onionMatches := onionRE.FindAll(buffer, -1)

	log.Printf("URLS : %v", onionMatches)

	onionMatches = removeDuplicates(onionMatches)

	log.Printf("URLS : %v", onionMatches)

	for i:=0; i < len(onionMatches); i++ {
		buffer = bytes.Replace([]byte(buffer), []byte(onionMatches[i]), []byte(string(onionMatches[i]) + domain), -1)
	}

	//jsRE := regexp.MustCompile("((S|s)(c|C)(r|R)(i|I)(p|P)(t|T))|((j|J)|(a|A)(v|V)(a|A))")

	//buffer = jsRE.ReplaceAll([]byte(buffer), []byte("disabled-js"))

	//log.Printf("[*] Resp Cookies: %v\n", resp.Header)

	// close the previous response body
	resp.Body.Close()

	newResp := resp

	// update the response content length
	newResp.ContentLength = int64(len(buffer))
	newResp.TransferEncoding = nil

	// assign the new response body as a io.ReadCloser
	newResp.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))

	// write the response to the source connection
	newResp.Write(sWriter)
	sWriter.Flush()
}

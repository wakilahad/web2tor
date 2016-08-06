package proxy

import (
	"io"
	"bytes"
	"time"
	"strconv"
	"golang.org/x/net/proxy"
	"regexp"
	"bufio"
	"net"
	"errors"
	"log"
	"net/http"
	"crypto/tls"
	"html/template"
	"net/http/httputil"
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

func (s *Server) ListenAndServe() error {

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

		go s.HandleConn(conn)
	}
}

func (s *Server) HandleConn(conn net.Conn) {

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
	log.Printf("Request: %v\n", req)

	// seperates out the host and URI
	httpRE,err := regexp.Compile("([a-zA-Z0-9]+.onion){1}.chr1s.co")
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
	log.Println("[*] Proxying URL : ", URL)

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

	// set a query string parameter to bust the cache
	q := req.URL.Query()
	q.Set("w2t", strconv.Itoa(int(time.Now().Unix())))
	req.URL.RawQuery = q.Encode()

	// DEBUG:
	log.Println("[*] Modified URL : ", req.URL.String())

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
	log.Println("Response Code : ", resp.Status)

	if len(resp.Header.Get("Location")) > 0 {
		location := bytes.Replace([]byte(resp.Header.Get("Location")), []byte(Host), []byte(Host + ".chr1s.co"), -1)
		resp.Header.Set("Location", string(location))
	}

	// parse the response and replace any anchor links with an updated URL
	buffer := make([]byte, 2^16)
	tmpNewReader := &bytes.Buffer{}

	for {
		n, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			break
		} else if err == io.EOF && n == 0 {
			break
		}

		// write to a temporary buffer
		tmpNewReader.Write(buffer[0:n])
	}

	// replace all host URL's with modified ones
	fromStrings := []string{
		Host,
	}

	toStrings := []string{
		Host + ".chr1s.co",
	}

	if len(fromStrings) != len(toStrings) {
		// PROGRAMMER ERROR: fix your from strings and to strings!
		panic("fix your from strings and to strings!")
	}

	// replace all the strings that need replacing in the returned document
	log.Printf("Original Length: %d\n", len(tmpNewReader.Bytes()))
	buffer = bytes.Replace(tmpNewReader.Bytes(), []byte(fromStrings[0]), []byte(toStrings[0]), -1)
	log.Printf("   New   Length: %d\n", len(buffer))

	/*
	// use regular expressions to replace all relative URI's with a modified version
	uriRE, err := regexp.Compile("((href=)|(src=))('|\")(http://)?(/)?([^@:]+)")
	if err != nil {
		panic("fix your regular expression : " + err.Error())
	}

	uriRE.ReplaceAllFunc(buffer, func (what []byte) []byte {
		matches := uriRE.FindSubmatch(what)

		if len(matches) != 7 {
			log.Println("[!] Error With RegEX : ", len(matches))
			return what
		}

		new := string(matches[0]) + string(matches[3]) + "http://159.203.57.234/" + Host + "/" + string(matches[6])

		log.Println("Replacing ", string(what), " with ", new)

		return []byte(new)
	})
	*/

	/*for i := 1; i < len(fromStrings); i++ {
		newBuffer := Replace(buffer, []byte(fromStrings[i]), []byte(toStrings[i]))
		buffer = newBuffer
		log.Printf("Length: %d bytes\n", len(buffer))
	}*/

	// create a new io.Reader with the modified document
	//tmp2NewReader := bytes.NewBuffer(buffer)

	// close the previous response body
	resp.Body.Close()

	// update the response content length
	resp.ContentLength = int64(len(buffer))
	resp.TransferEncoding = nil

	dump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		log.Printf("[!] Error Dumping Response : %s\n", err.Error())
		return
	}

	log.Printf("Precursor: %s\n", dump)
	log.Printf("Body Len : %s\n", len(buffer))

	sWriter.Write(dump)
	sWriter.Write(buffer)

	// assign the new response body as a io.ReadCloser
	//resp.Body = ioutil.NopCloser(tmp2NewReader)

	// write the response to the source connection
	//resp.Write(sWriter)
	sWriter.Flush()
}

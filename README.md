# web2tor

Web2Tor is a HTTP proxy written entirely in Go that forwards all HTTP and HTTPS connections to a TOR SOCKS5 client running on port 127.0.0.1:9050 (the default).  It takes as a command line parameter the domain extension to be appended to each .onion URL to access the TOR network. 

The software must also be started as root, because it by default binds to ports 80 (HTTP) and 443 (HTTPS).  It immediately drops privileges to UID/GID 1000 before hosting any web content, however.

# example

    ```
    tor &
    sudo ./web2tor yourdomain.com
    ```

This starts the TOR client, then starts the HTTP proxy.  You'll then be able to access the TOR network at [tor url.onion].yourdomain.com, once you set up a wildcard A record at your DNS provider like so:

    ```
    <yourdomain.com DNS config>

    *   <server IP>
    ```

# usage

Copyright (c) 2016 Chris Pergrossi


Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
    
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


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

# web2tor

Web2Tor is a HTTP proxy written entirely in Go that forwards all HTTP and HTTPS connections to a TOR SOCKS5 client running on port 127.0.0.1:9050 (the default).  It takes as a command line parameter the domain extension to be appended to each .onion URL to access the TOR network.

The software must also be started as root, because it by default binds to ports 80 (HTTP) and 443 (HTTPS).  It immediately drops privileges to UID/GID 1000 before hosting any web content, however.

# example

Remember! The domain name must begin with a period in order for domains like blahblahblah.onion.yourdomain.com to work, or else only blahblahblah.onionyourdomain.com will work.

    tor &
    sudo ./web2tor .yourdomain.com

This starts the TOR client, then starts the HTTP proxy.  You'll then be able to access the TOR network at [tor url.onion].yourdomain.com, once you set up a wildcard A record at your DNS provider like so:

    <yourdomain.com DNS config>

    *   <server IP>

Here is the full example I run on my perseonal server:

    tor &
    sudo ./web2tor .to.chr1s.co

    // DNS Entry for chr1s.co

    *   .to     111.111.111.111

# HTTPS support

By default the server runs on ports 80 and 443 with HTTP and HTTPS connections.  The server needs a server.pem (certificate) and a server.key (private key) in the same folder as the executable in order to start the HTTPS server.  You'll need a wildcard certificate (between 50 - 100$) to support all onion URL's on your server (or you can reprogram it to use a training URI but I much prefer the subdomain mechanism), however, I find the following mechanism the best:

LetsEncrypt is a semi-automated method of issuing free fully validated certificates to domains using an easy to install and run program (just apt-get install letsencrypt and follow the instructions).  So there are two ways you can use this:  you either know which domains you really want secured beforehand (say you only use your proxy for certain websites) or you can save the URL's people go to on your server and then write a small script to issue a certificate for all those domains automatically when you feel like updating your certificate.  I warn you, you can only reissue a certificate for a certain domain I think maybe 5 times a week or similar, so make sure you have the right script to get the domains correct or you'll have to wait to issue a corrected certificate.

    # shutdown the proxy beforehand, as this uses port 80/443 itself
    letsencrypt certonly --standalone -d www.chr1s.co -d chr1s.co -d blahblah.onion.to.chr1s.co     # etc

# usage

Copyright (c) 2016 Chris Pergrossi


Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

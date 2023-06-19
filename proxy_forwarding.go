package pineweb

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

func (pinehost *Pinehost) webuiforwarding() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		//If we have no pinecone url, display the index
		if len(req.URL.Path) <= 1 {
			// "" or "/"
			data := pinehost.peers()
			makeIndexPage(&data).WriteTo(w)
			return
		}
		if len(req.URL.Path) < 65 { //slash+pineaddress
			fmt.Fprintf(w, "<html><body><h1>Pineaddress too short</h1><p><a href=\"./\">Return</a></p></body></html>")
			return
		}

		// We cant loopback dial with quic, so we directly go to our local server
		if pinehost.addr == req.URL.Path[1:65] {
			http.Redirect(w, req, pinehost.lhost, http.StatusSeeOther)
			return
		}

		http.Redirect(w, req, "http://"+req.URL.Path[1:65], http.StatusSeeOther)
	}
}

func (pinehost *Pinehost) handleProxy(transport http.RoundTripper) http.HandlerFunc {
	// I am not quite sure if there is not a more straight forward way,
	// like just invoking the round trip? This certainly works tho.
	proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {}}
	proxy.Transport = transport

	return func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	}
}

func (pinehost *Pinehost) SpawnForwardingProxy(proxy string) net.Addr {
	// Proxy
	mux := http.NewServeMux()
	mux.HandleFunc("/", pinehost.handleProxy(pinehost.pineweb.Client().Transport))
	listener, err := net.Listen("tcp", proxy)
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go srv.Serve(listener)

	return listener.Addr()
}

func (pinehost *Pinehost) SpawnWebUI(webui string) net.Addr {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pinehost.webuiforwarding())

	listener, err := net.Listen("tcp", webui)
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, mux)

	return listener.Addr()
}

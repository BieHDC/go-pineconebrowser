package pineweb

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

func pinewebProxy(serverurl string, transport http.RoundTripper) http.HandlerFunc {
	replacee, err := url.Parse("http://" + serverurl)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(replacee)
	proxy.Transport = transport

	return func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	}
}

var openconns = map[string]net.Addr{} //maybe sync.Map
func spawnTempLocalServer(peerid string, transport http.RoundTripper) net.Addr {
	cached, exists := openconns[peerid]
	if exists {
		return cached
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	openconns[peerid] = listener.Addr()

	mux := http.NewServeMux()
	mux.HandleFunc("/", pinewebProxy(peerid, transport))

	var clients atomic.Int64
	var srv http.Server
	srv = http.Server{
		Handler: mux,
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				clients.Add(1)
			case http.StateHijacked, http.StateClosed:
				fin := clients.Add(-1)
				if fin == 0 {
					delete(openconns, peerid)
					listener.Close()
					srv.Shutdown(context.Background())
				}
			}
		},
		// TODO: The issue is the following:
		// If we dont timeout, we keep taking up slots on the remote server
		// If we do timeout, its on the client to ether ping once in a while
		// or the connection will disappear, shutting the website down from
		// the clients perspective.
		// We could ether track open proxys, or somehow always use the same
		// local port for the remote host, and just respawn the server.
		// For now thats out of scope and the server can always kick the
		// sleeping client on its side.
		IdleTimeout: 60 * time.Minute,
	}

	go srv.Serve(listener)

	return listener.Addr()
}

func (pinehost *Pinehost) webuiserverproxy() http.HandlerFunc {
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

		localwebserver := spawnTempLocalServer(req.URL.Path[1:65], pinehost.pineweb.Client().Transport)
		http.Redirect(w, req, "http://"+localwebserver.String(), http.StatusSeeOther)
	}
}

func (pinehost *Pinehost) SpawnProxy(webui string) net.Addr {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pinehost.webuiserverproxy())

	listener, err := net.Listen("tcp", webui)
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, mux)

	return listener.Addr()
}

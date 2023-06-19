package pineweb

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	pineconeRouter "github.com/matrix-org/pinecone/router"
	pineconeEvents "github.com/matrix-org/pinecone/router/events"
	pineconeSessions "github.com/matrix-org/pinecone/sessions"
)

type pinehostAttributes struct {
	Name       string
	IsPinehost bool
	Webhoster  bool
	Version    string
	Endpoints  string
}

type Pinehost struct {
	logger   *log.Logger
	pinehost *pineconeSessions.HTTP
	pineweb  *pineconeSessions.HTTP
	router   *pineconeRouter.Router
	addr     string //needed for proxy loopback
	lhost    string //needed for proxy loopback
	pa       pinehostAttributes

	//internal
	knownhosts sync.Map
	endpoints  map[string]*endpoint
}

func PinehostSpawn(logger *log.Logger, router *pineconeRouter.Router, addr string, username string, serverurl string) *Pinehost {
	var pa pinehostAttributes
	pa.Name = username
	pa.IsPinehost = true
	if serverurl != "NONE" {
		pa.Webhoster = true
	}
	pa.Version = "0.8"

	protocols := []string{"pinehost", "pineweb"}
	dialer := pineconeSessions.NewSessions(logger, router, protocols)

	pinehost := &Pinehost{
		logger:   logger,
		pinehost: dialer.Protocol("pinehost").HTTP(),
		pineweb:  dialer.Protocol("pineweb").HTTP(),
		router:   router,
		addr:     addr,
		lhost:    serverurl,
		pa:       pa,
	}

	pinehost.setupHandlers()

	if pa.Webhoster {
		pinehost.pineweb.Mux().HandleFunc("/", handleProxyRequest(serverurl)) //for remote-to-local forwarding
	}

	go pinehost.eventWatch()

	// For the list in the webui
	paself := pinehost.pa
	paself.Name = pa.Name + " (You)"
	pinehost.knownhosts.Store(addr, &paself)

	return pinehost
}

func (pinehost *Pinehost) eventWatch() {
	eventchannel := make(chan pineconeEvents.Event)
	pinehost.router.Subscribe(eventchannel)
	pinehost.logger.Print("Pinehost Event Listener active")
	for event := range eventchannel {
		switch e := event.(type) {
		case pineconeEvents.PeerAdded:
			pinehost.logger.Print("Pinehost: ", PineconeEventToString(event), " received from: ", e.PeerID)
			go pinehost.checkForPinehost(e.PeerID)

		case pineconeEvents.PeerRemoved:
			pinehost.logger.Print("Pinehost: ", PineconeEventToString(event), " received from: ", e.PeerID)
			pinehost.knownhosts.Delete(e.PeerID)
		default:
		}
	}
}

func (pinehost *Pinehost) checkForPinehost(peerid string) {
	// fixme needs a delay otherwise drops connection
	// but also 2 devices should talk to each other at the very same point
	// needs some investigation, because its still wonky
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration((rand.Intn(5000) + 1000)) * time.Millisecond)

	remotePinehost := &pinehostAttributes{}

	if pinehost.remoteIsInstance(peerid) {
		remotePinehost.IsPinehost = true

		rname, err := pinehost.RemoteGet(peerid, pinehost.GetQuery("username"))
		if err == nil {
			remotePinehost.Name = rname
		} else {
			pinehost.logger.Println(err.Error())
			remotePinehost.Name = "FAILED:NAME"
		}

		rversion, err := pinehost.RemoteGet(peerid, pinehost.GetQuery("version"))
		if err == nil {
			remotePinehost.Version = rversion
		} else {
			pinehost.logger.Println(err.Error())
			remotePinehost.Version = "FAILED:VERSION"
		}

		rendpoints, err := pinehost.RemoteGet(peerid, pinehost.GetQuery("endpoints"))
		if err == nil {
			remotePinehost.Endpoints = rendpoints
		} else {
			pinehost.logger.Println(err.Error())
			remotePinehost.Version = "FAILED:ENDPOINTS"
		}

		if strings.Contains(rendpoints, "webhoster") {
			remotePinehost.Webhoster = true
		}
	}

	pinehost.knownhosts.Store(peerid, remotePinehost)
}

func (pinehost *Pinehost) remoteIsInstance(addr string) bool {
	// fixme due to wonkyness mentioned in caller of this,
	// lets force retry a few times. which also means we
	// are going to annoy any non-pinehost peers
	for i := 0; i < 5; i++ {
		resp, err := pinehost.RemoteGet(addr, pinehost.GetQuery("isinstance"))
		if err != nil {
			// retry after delay
			rand.Seed(time.Now().UnixNano())
			time.Sleep(time.Duration((rand.Intn(2000) + 1000)) * time.Millisecond)
			continue
		}
		if resp != "Pinehost Instance" {
			//fmt.Printf("expected >Pinehost Instance<, got >%s<\n", string(resp)) //debug
			return false
		}
		return true //is a pinehost instance
	}
	//pinehost.logger.Printf("call failed for %s\n", addr) //debug
	return false //is not a pinehost instance
}

func handleProxyRequest(serverurl string) http.HandlerFunc {
	replacee, err := url.Parse(serverurl)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(replacee)

	return func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	}
}

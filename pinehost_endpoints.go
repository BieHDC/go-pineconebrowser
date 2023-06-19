package pineweb

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type endpoint struct {
	Query string
	Reply func() string
}

func (pinehost *Pinehost) setupHandlers() {
	pinehost.endpoints = map[string]*endpoint{
		"isinstance": {"/", func() string { return "Pinehost Instance" }},
		"username":   {"/username", func() string { return pinehost.pa.Name }},
		"version":    {"/version", func() string { return pinehost.pa.Version }},
		"endpoints":  {"/endpoints", func() string { return pinehost.pa.Endpoints }},
	}
	if pinehost.pa.Webhoster {
		pinehost.endpoints["webhoster"] = &endpoint{"/webhoster", func() string {
			return "fixme what could or should we return here? might be interesting if we get a query details page that also holds other things"
		}}
	}

	var registeredEnpoints []string
	for k, ep := range pinehost.endpoints {
		ep := ep //you need this
		registeredEnpoints = append(registeredEnpoints, k)
		pinehost.pinehost.Mux().HandleFunc(ep.Query, func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, ep.Reply())
		})
	}

	sort.Slice(registeredEnpoints, func(i, j int) bool {
		return registeredEnpoints[i] < registeredEnpoints[j]
	})
	pinehost.pa.Endpoints = strings.Join(registeredEnpoints, ",")
}

func (pinehost *Pinehost) GetQuery(q string) string {
	query, exists := pinehost.endpoints[q]
	if !exists {
		panic("you tried to query a non existend endpoint")
	}
	return query.Query
}

func (pinehost *Pinehost) RemoteGet(addr, query string) (string, error) {
	resp, err := pinehost.call(addr, query)
	return string(resp), err
}

func (pinehost *Pinehost) call(addr, query string) ([]byte, error) {
	resp, err := pinehost.pinehost.Client().Get("http://" + addr + query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

package pineweb

import (
	"bytes"
	"html/template"
	"strconv"

	pineconeRouter "github.com/matrix-org/pinecone/router"
)

// Makes a nice list for the index site
type Peerparsed struct {
	Key   string
	Ports []string
	Zone  string
	PA    pinehostAttributes
}

func (pinehost *Pinehost) peers() []Peerparsed {
	peers := getPeers(pinehost.router)
	for i, peer := range peers {
		valr, exists := pinehost.knownhosts.Load(peer.Key)
		if !exists {
			peers[i].PA.Name = "IMPOSSIBLE YOU NEED A NAME"
			continue
		}
		val, ok := valr.(*pinehostAttributes)
		if !ok {
			peers[i].PA.Name = "IMPOSSIBLE YOU HAVE TO BE A PinehostAttributes"
			continue
		}
		peers[i].PA = *val
	}
	return peers
}

func getPeers(router *pineconeRouter.Router) []Peerparsed {
	peers := router.Peers()
	var peerslice []Peerparsed
	for _, peer := range peers {
		exists := false
		for j, peerparsed := range peerslice {
			if peerparsed.Key == peer.PublicKey {
				//if we already have an entry, append and replace information
				exists = true
				peerslice[j].Ports = append(peerslice[j].Ports, strconv.Itoa(peer.Port))
				if peer.Zone != "" {
					peerslice[j].Zone = peer.Zone
				}
			}
		}
		// dont replace the original entry if it did already exist, we appended to it
		if !exists {
			peerslice = append(peerslice, Peerparsed{Key: peer.PublicKey, Ports: []string{strconv.Itoa(peer.Port)}, Zone: peer.Zone})
		}
	}
	return peerslice
}

func makeIndexPage(pi *[]Peerparsed) *bytes.Buffer {
	indextemplate := template.Must(template.New("index").Parse(tmplt))

	page := &bytes.Buffer{}
	err := indextemplate.Execute(page, pi)
	if err != nil {
		panic(err)
	}

	return page
}

const tmplt = `<html>
  <head>
    <style>
      html * {
        font-family: monospace;
      }
      html, body {
        height: 100%;
      }
      html {
        display: table;
        margin: auto;
      }
      body {
        display: table-cell;
        vertical-align: middle;
      }
      table, th, td {
        border:1px solid black;
        padding: 5px;
      }
    </style>
  </head>
  <body>
    <table>
      <tr>
        <th>Public Key</th>
        <th>Port(s)</th>
        <th>Zone</th>
        <th>Name</th>
        <th>Webhoster</th>
        <th>Supported Endpoints</th>
      </tr>
      {{ range . }}
      <tr>
        {{ if .PA.Webhoster }}
        <td><a href="{{ .Key }}" target="_blank" rel="noopener noreferrer">{{ .Key }}</a></td>
        {{ else }}
        <td>{{ .Key }}</td>
        {{ end }}
        <td>{{ .Ports }}</td>
        <td>{{ .Zone }}</td>
        {{ if .PA.IsPinehost }}
        <td>{{ .PA.Name }}</td>
        <td>{{ .PA.Webhoster }}</td>
        <td>{{ .PA.Endpoints }}</td>
        {{ else }}
        <td>Not a Pinehost</td>
        <td></td>
        <td></td>
        {{ end }}
      </tr>
      {{ end }}
    </table>
  </body>
</html>
`

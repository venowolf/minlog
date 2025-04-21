package global

import (
	"os"
	"strings"
)

// store global settings, including namespace, nodename, k8s-dns, loki-endpoint, etc..
type GlobalSettings struct {
	Namespaces []string
	NodeName   string
	//Dns        string //TODO
	Lokiep string // default value is set at parseflags
}

func NewGSettings(nn, ns, lokiep string) *GlobalSettings {
	// initialize namespaces, grafana-agent-flow will push those pods which running in these namespaces to lokiep
	gs := &GlobalSettings{
		Lokiep: lokiep,
	}
	if nn == "" || nn == "all" {
		gs.Namespaces = make([]string, 0)
	} else {
		gs.Namespaces = strings.Split(nn, ",")
	}

	// node name, if not specified, get it from os.Getenv("HOSTNAME")
	if nn == "" {
		gs.NodeName = os.Getenv("HOSTNAME")
	} else {
		gs.NodeName = nn
	}

	return gs
}

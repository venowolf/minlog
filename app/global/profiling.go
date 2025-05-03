package global

import (
	"os"
	"strings"
)

var RiverFile string

// grafana-agent-flow url, default http://127.0.0.1:12345
var Gafurl string = "http://127.0.0.1:12345"

// store global settings, including namespace, nodename, k8s-dns, loki-endpoint, etc..
type GlobalSettings struct {
	NameSpaces []string

	NodeName string
	//NodeNameWithOutDash=strings.ReplaceAll(NodeName, "-", "")
	NodeNameWithOutDash string

	//Dns        string //TODO
	Lokiep string // default value is set at parseflags

	//container log dir
	AppLogs string
}

func NewGSettings(nn, ns, logdir, lokiep string) *GlobalSettings {
	// initialize namespaces, grafana-agent-flow will push those pods which running in these namespaces to lokiep
	gs := &GlobalSettings{
		Lokiep: lokiep,
	}
	if ns == "" || ns == "all" {
		gs.NameSpaces = make([]string, 0)
	} else {
		gs.NameSpaces = strings.Split(ns, ",")
	}

	// node name, if not specified, get it from os.Getenv("HOSTNAME")
	if nn == "" {
		gs.NodeName = os.Getenv("HOSTNAME")
	} else {
		gs.NodeName = nn
	}
	if logdir == "" {
		gs.AppLogs = "/var/log/containers"
	} else {
		gs.AppLogs = logdir
	}

	gs.NodeNameWithOutDash = strings.ReplaceAll(gs.NodeName, "-", "")
	return gs
}

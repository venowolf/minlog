package global

import (
	"os"
	"strings"
)

// global setttings, initialize in argparse or mod initializing
var AlloyFile string
var GAlloyurl string
var applogsdir string
var Lokiep string

func init() {
	/* default value is from global environment variables which defined in dockerfile
	Gafurl = os.Getenv("ALLOYURL")
	Lokiep = os.Getenv("LOKIEP")
	*/
	applogsdir = GetFromEnvOrDefaultValue("APPLOGSDIR")
}

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

func GetFromEnvOrDefaultValue(key string) string {
	value := os.Getenv(key)
	if value == "" {
		switch key {
		case "ALLOYURL":
			value = "http://127.0.0.1:12345"
		case "LOKIEP":
			value = "http://loki:3100/loki/api/v1/push"
		case "APPLOGSDIR":
			value = "/var/log/containers"
		case "ALLOYFILE":
			value = "/app/confs/grafana-alloy.alloy"
		}
	}
	return value
}

func NewGSettings(nn, ns string) *GlobalSettings {
	// initialize namespaces, grafana-alloy will push those pods which running in these namespaces to lokiep
	gs := &GlobalSettings{
		Lokiep:  Lokiep,
		AppLogs: applogsdir,
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

	gs.NodeNameWithOutDash = strings.ReplaceAll(gs.NodeName, "-", "")
	return gs
}

package agentflow

import (
	"os"
)

var nn string
var lokiep string

func init() {
	nn = os.Getenv("HOSTNAME")
}

func SetNodeName(name string) {
	nn = name
}

func GetNodeName() string {
	return nn
}

func initLokiep(f func() (string, error)) bool {
	l, e := f()
	if e != nil {
		return false
	} else {
		lokiep = l
	}
	return true
}
func GetLokiep() string {
	return lokiep
}

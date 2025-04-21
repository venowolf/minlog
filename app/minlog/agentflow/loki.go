package agentflow

import (
	"venomouswolf/minlog/app/log"
)

// update loki configure file loki.river

var lokiPort int = 12345

type loki struct {
	river   string
	appLogs string
}

func NewLoki(river string, appLogs string) *loki {
	l := log.GetLogger()
	l.Info("Loki client created")
	return &loki{
		river:   river,
		appLogs: appLogs,
	}
}

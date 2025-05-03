package agentflow

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"venomouswolf/minlog/app/global"

	"k8s.io/klog/v2"
)

// gaf --> grafana agent flow
// update loki configure file loki.river

type Gaf interface {
	Reload() error
}

type gaf struct {
	log klog.Logger

	//ready check, retry times, default 5 * 15 time.Second
	readycheck int
}

func NewGaf() Gaf {
	g := &gaf{
		readycheck: 5,
		log:        klog.Background().WithName("loki-server"),
	}
	g.log.Info("Creating loki-server object")
	return g
}

func (g *gaf) Reload() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	readyYes := false

	readyr, _ := http.NewRequest("GET", global.Gafurl+"/-/ready", strings.NewReader(""))
	readyr.Header.Set("Content-Type", "application/json")
	for i := 0; i < g.readycheck; i++ {
		if readyresp, e := client.Do(readyr); e == nil {
			rb := make([]byte, 0, 64)
			if _, e = readyresp.Body.Read(rb); e == nil {
				if string(rb) == "Agent is ready." {
					readyYes = true
					break
				}
			}

		} else {
			time.Sleep(15 * time.Second)
		}
	}
	if !readyYes {
		err := errors.New("Grafana-agent-flowError")
		g.log.Error(err, "Grafana agent flow is not ready")
		return err
	}

	req, _ := http.NewRequest("POST", global.Gafurl+"/-/reload", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	//req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})

	resp, err := client.Do(req)
	if err == nil {
		rb := make([]byte, 0, 32)
		if _, err = resp.Body.Read(rb); err == nil {
			if string(rb) == "config reloaded" {
				g.log.Info("Request reloading ... Done")
				return err
			}
		}
	}
	g.log.Error(err, "Request reloading ... Failed")
	return err
}

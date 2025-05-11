package grafanaalloy

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"venowolf/minlog/app/global"

	"k8s.io/klog/v2"
)

// galloy --> grafana alloy
// update grafana alloy configure file config.alloy

type GAlloy interface {
	Reload() error
}

type galloy struct {
	log klog.Logger

	//ready check, retry times, default 5 * 15 time.Second
	readycheck int
}

func NewGAlloy() GAlloy {
	g := &galloy{
		readycheck: 5,
		log:        klog.Background().WithName("grafana-alloy"),
	}
	g.log.Info("Creating grafana alloy wrapper object")
	return g
}

func (g *galloy) Reload() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	readyYes := false

	readyr, _ := http.NewRequest("GET", global.GAlloyurl+"/-/ready", strings.NewReader(""))
	readyr.Header.Set("Content-Type", "application/json")
	for i := 0; i < g.readycheck; i++ {
		if readyresp, e := client.Do(readyr); e == nil {
			rb := make([]byte, 0, 64)
			if _, e = readyresp.Body.Read(rb); e == nil {
				if string(rb) == "Alloy is ready." {
					readyYes = true
					break
				}
			}

		} else {
			time.Sleep(15 * time.Second)
		}
	}
	if !readyYes {
		err := errors.New("Grafana-alloyError")
		g.log.Error(err, "Grafana alloy  is not ready")
		return err
	}

	req, _ := http.NewRequest("POST", global.GAlloyurl+"/-/reload", strings.NewReader(""))
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

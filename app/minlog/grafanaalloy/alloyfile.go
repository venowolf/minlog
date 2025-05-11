package grafanaalloy

import (
	"bytes"
	"os"
	"text/template"
	"venowolf/minlog/app/global"

	"k8s.io/klog/v2"
)

type AlloyFile interface {
	GenAlloyFile(podsm map[string]*PodInfo) error
}

type alloyfile struct {
	gsettings *global.GlobalSettings

	alloy_template   *template.Template
	podconf_template *template.Template

	alloybuffer *bytes.Buffer

	alloyFile string

	log klog.Logger
}

func NewAlloyFile(gs *global.GlobalSettings) AlloyFile {
	return &alloyfile{
		alloyFile:        global.AlloyFile,
		gsettings:        gs,
		alloy_template:   template.Must(template.New("alloy_template").Parse(alloy_template)),
		podconf_template: template.Must(template.New("podconf_template").Parse(pod_template)),
		alloybuffer:      bytes.NewBuffer(nil),
		log:              klog.Background().WithName("grafana-alloy"),
	}
}

type PodInfo struct {
	PodName string
	// Containers key: container name, value: container id
	ContainerMap map[string]string
	NameSpace    string
	ServiceName  string
	AppLogs      string
}

type PodInfo_Tmpl struct {
	Pod                 *PodInfo
	NodeNameWithOutDash string
}

func (r *alloyfile) writeLokiAlloy() error {
	return r.alloy_template.Execute(r.alloybuffer, r.gsettings)
}

func (r *alloyfile) writePodAlloy(pod *PodInfo) error {
	tpod := PodInfo_Tmpl{
		NodeNameWithOutDash: r.gsettings.NodeNameWithOutDash,
		Pod:                 pod,
	}
	return r.podconf_template.Execute(r.alloybuffer, tpod)
}

func (r *alloyfile) GenAlloyFile(podsm map[string]*PodInfo) error {
	// write loki river
	if err := r.writeLokiAlloy(); err != nil {
		r.log.Error(err, "Error to gen loki settings")
		return err
	}
	// write pods river
	for _, pod := range podsm {
		if err := r.writePodAlloy(pod); err != nil {
			r.log.Error(err, "Error to gen river block")
			return err
		}
	}

	return r.writeToFile()
}
func (r *alloyfile) writeToFile() error {
	if f, err := os.OpenFile(r.alloyFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		if l, e := f.Write(r.alloybuffer.Bytes()); l != r.alloybuffer.Len() || e != nil {
			r.log.Error(err, "Fail to create "+r.alloyFile)
			f.Close()
			return e
		} else {
			r.log.Info("Done to create " + r.alloyFile)
			return f.Close()
		}
	} else {
		return err
	}
}

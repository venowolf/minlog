package agentflow

import (
	"bytes"
	"os"
	"text/template"
	"venomouswolf/minlog/app/global"

	"k8s.io/klog/v2"
)

type River interface {
	GenRiver(podsm map[string]*PodInfo) error
}

type river struct {
	gsettings *global.GlobalSettings

	flowconf_template *template.Template
	podconf_template  *template.Template

	riverbuffer *bytes.Buffer

	riverFile string

	log klog.Logger
}

func NewRiver(gs *global.GlobalSettings) River {
	return &river{
		riverFile:         global.RiverFile,
		gsettings:         gs,
		flowconf_template: template.Must(template.New("flowconf_template").Parse(agentflow_template)),
		podconf_template:  template.Must(template.New("podconf_template").Parse(pod_template)),
		riverbuffer:       bytes.NewBuffer(nil),
		log:               klog.Background().WithName("grafana-agent-flow-river"),
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

func (r *river) writeLokiRiver() error {
	return r.flowconf_template.Execute(r.riverbuffer, r.gsettings)
}

func (r *river) writePodRiver(pod *PodInfo) error {
	tpod := PodInfo_Tmpl{
		NodeNameWithOutDash: r.gsettings.NodeNameWithOutDash,
		Pod:                 pod,
	}
	return r.podconf_template.Execute(r.riverbuffer, tpod)
}

func (r *river) GenRiver(podsm map[string]*PodInfo) error {
	// write loki river
	if err := r.writeLokiRiver(); err != nil {
		r.log.Error(err, "Error to gen loki settings")
		return err
	}
	// write pods river
	for _, pod := range podsm {
		if err := r.writePodRiver(pod); err != nil {
			r.log.Error(err, "Error to gen river block")
			return err
		}
	}

	return r.writeToFile()
}
func (r *river) writeToFile() error {
	if f, err := os.OpenFile(r.riverFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		if l, e := f.Write(r.riverbuffer.Bytes()); l != r.riverbuffer.Len() || e != nil {
			r.log.Error(err, "Fail to create "+r.riverFile)
			f.Close()
			return e
		} else {
			r.log.Info("Done to create " + r.riverFile)
			return f.Close()
		}
	} else {
		return err
	}
}

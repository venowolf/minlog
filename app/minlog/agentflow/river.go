package agentflow

import (
	"bytes"
	"os"
	"text/template"
	"venomouswolf/minlog/app/global"
)

type river struct {
	gsettings *global.GlobalSettings

	flowconf_template *template.Template
	podconf_template  *template.Template


	riverbuffer *bytes.Buffer
}

func NewRiver() *river {
	return &river{
		gsettings:         &global.GSettings,
		flowconf_template: template.Must(template.New("flowconf_template").Parse(agentflow_template)),
		podconf_template:  template.Must(template.New("podconf_template").Parse(pod_template)),
		riverbuffer: 	 bytes.NewBuffer(nil),
	}
}



type TPodInfo struct {
	Podname       string
	Containername string
	Containerid   string
	Namespace     string
	Serviceename  string
}



func (r *river) writeLokiRiver() error {
	return r.flowconf_template.Execute(r.riverbuffer, global.GSettings)
}

func (r *river) writePodsRiver(pods []) error {
	t := template.Must(template.New("agent.river").Parse(pod_template))
	podInfo := TPodInfo{
		Podname:       "podname",
		Containername: "containername",
		Containerid:   "containerid",
		Namespace:     "namespace",
		Serviceename:  "serviceename",
	}
	t.Execute(os.Stdout, podInfo)
	return nil
}

func (r *river) writePodRiver(pod TPodInfo) error{
	return r.podconf_template.Execute(r.riverbuffer, pod)
}

func (r *river) GenRiver() error {
	// write loki river
	if err := r.writeLokiRiver(); err != nil {
		return err
	}
	// write pods river
	if err := r.writePodsRiver(); err != nil {
		return err
	}
	return nil
}



package k8s

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

// return the name of the pod's owerreferrence
func getServiceNameOfPod(p *v1.Pod) string {
	refkind := p.OwnerReferences[0].Kind
	switch refkind {
	case "StatefulSet", "Node", "DaemonSet":
		return p.OwnerReferences[0].Name
	case "ReplicaSet": //kube-apiserver-node
		rsname := p.OwnerReferences[0].Name
		idx := strings.LastIndex(rsname, "-")
		return rsname[0:idx]
	}
	return ""
}

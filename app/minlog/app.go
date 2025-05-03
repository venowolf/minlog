package minlog

import (
	"k8s.io/client-go/kubernetes"
)

type minLog struct {
	kclient *kubernetes.Clientset
}

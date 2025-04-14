package minlog

import (
	"venomouswolf/minlog/app/log"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

var logger *zap.Logger = log.GetLogger()

type minLog struct {
	kclient *kubernetes.Clientset
}

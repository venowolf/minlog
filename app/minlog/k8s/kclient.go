package k8s

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
	"venomouswolf/minlog/app/global"
	"venomouswolf/minlog/app/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/client-go/tools/clientcmd"
)

type KClient interface {
	Start(ctx context.Context) error
	Profilling()
}

func NewKClient(nn, ns, lokiep string, runningOnly bool) KClient {
	// creates the out-cluster config
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")

	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	if err != nil {
		log.GetLogger().Panic("failed to create in-cluster config", zap.String("FatalError", err.Error()))
		return nil
	}
	kc := &kclient{
		gsettings: global.NewGSettings(nn, ns, lokiep),
		zlog:      log.GetLogger(),

		clientset:   kubernetes.NewForConfigOrDie(config),
		runningOnly: runningOnly,
		podmaps:     make(map[string]string),
		pqueue:      workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]()),
	}

	if nn != "" {
		kc.gsettings.NodeName = nn
	} else {
		lo := metav1.ListOptions{
			FieldSelector: "metadata.name=" + os.Getenv("HOSTNAME"),
		}
		selfp, err := kc.clientset.CoreV1().Pods("").List(context.TODO(), lo)
		if len(selfp.Items) != 1 || err != nil {
			return nil
		}
		kc.gsettings.NodeName = selfp.Items[0].Spec.NodeName
	}

	if ns != "all" && ns != "" {
		kc.gsettings.Namespaces = strings.Split(ns, ",")
	}

	return kc
}

type kclient struct {
	gsettings *global.GlobalSettings
	zlog      *zap.Logger

	clientset *kubernetes.Clientset

	runningOnly bool

	// pods keeps all the pods which are running on this node(pods logs will be pushed to loki)
	// key: namesapace/podname
	// value: pod's object
	podmaps map[string]string

	pqueue workqueue.TypedRateLimitingInterface[string]
}

func (kc *kclient) startWithNamespace(ctx context.Context, nn string) error {
	fs := fields.Set{"spec.nodeName": kc.gsettings.NodeName}
	if kc.runningOnly {
		fs["status.phase"] = "Running"
	}
	podInformer := cache.NewSharedInformer(
		cache.NewListWatchFromClient(kc.clientset.CoreV1().RESTClient(), "pods", nn, fs.AsSelector()),
		&v1.Pod{},
		1*time.Minute)
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				kc.pqueue.Add(fmt.Sprintf("A#%s", key))
				kc.zlog.Info("Pod added", zap.String("namespace/name", key))
			}
		},
		UpdateFunc: func(oobj, nobj interface{}) {
			if oobj.(*v1.Pod).ResourceVersion == nobj.(*v1.Pod).ResourceVersion {
				return // Periodic resync will send update events for all known pods.
			}
			key, err := cache.MetaNamespaceKeyFunc(nobj)
			if err == nil {
				kc.pqueue.Add(fmt.Sprintf("U#%s", key))
				kc.zlog.Info("Pod updated", zap.String("namespace/name", key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				kc.pqueue.Add(fmt.Sprintf("D#%s", key))
				kc.zlog.Info("Pod deleted", zap.String("namespace/name", key))
			}
		},
	})
	go kc.worker(ctx)
	return nil
}
func (kc *kclient) Start(ctx context.Context) error {
	if len(kc.gsettings.Namespaces) == 0 {
		kc.startWithNamespace(ctx, "")
	} else {
		for _, ns := range kc.gsettings.Namespaces {
			kc.startWithNamespace(ctx, ns)
		}
	}
}

func (kc *kclient) worker(ctx context.Context) {
	for kc.processNextWorkItem(ctx) {
	}
}

func (kc *kclient) processNextWorkItem(ctx context.Context) bool {
	key, quit := kc.pqueue.Get()
	if quit {
		return false
	}
	defer kc.pqueue.Done(key)

	return kc.sync(ctx, key)
}

func (kc *kclient) sync(ctx context.Context, key string) bool {
	nnkey := key[3:]
	ns, name, err := cache.SplitMetaNamespaceKey(nnkey)
	// string convert to
	if err != nil {
		kc.zlog.Error("Failed to split meta namespace cache key", zapcore.Field{Type: zapcore.ErrorType, Key: nnkey})
		return false
	}

	startTime := time.Now()
	kc.zlog.Info("Parse key", zap.String("namespace/name", nnkey), zap.Time("startTime", startTime))

	defer func() {
		kc.zlog.Info("Finished", zap.String("namespace/name", nnkey), zap.Time("startTime", time.Now()))
	}()
	switch key[0:1] {
	case "A", "U":
		pod, e := kc.clientset.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if e != nil {
			kc.zlog.Error("Failed to split meta namespace cache key", zapcore.Field{Type: zapcore.ErrorType, Key: key[3:]})
			return false
		}

		shas := make([]string, len(pod.Status.ContainerStatuses))
		for _, container := range pod.Status.ContainerStatuses {
			cid := strings.Split(container.ContainerID, "//")
			shas = append(shas, cid[1])
		}
		slices.Sort(shas)

		msha, inew := kc.podmaps[nnkey]
		if inew {
			if msha != strings.Join(shas, ",") {
				inew = false
			}
		}
		if !inew {
			kc.podmaps[nnkey] = strings.Join(shas, ",")
			//TODO: Generating a loki config.river, then curl http://grafana-agent-flow:12345/-/reload
			//minlog.
		}
	case "D":
		log.GetLogger().Info("Pod deleted", zap.String("namespace/name", nnkey))
	}

	return true
}

func (kc *kclient) profillingWithNamespace(nn string) {
	fs := fields.Set{"spec.nodeName": kc.gsettings.NodeName}
	if kc.runningOnly {
		fs["status.phase"] = "Running"
	}
	pods, err := kc.clientset.CoreV1().Pods(nn).List(context.TODO(), metav1.ListOptions{FieldSelector: fs.AsSelector().String()})
	if err != nil {
		kc.zlog.Error("Failed to list pods", zapcore.Field{Type: zapcore.ErrorType, Key: err.Error()})
		return
	}
	for _, pod := range pods.Items {
		if key, err := cache.MetaNamespaceKeyFunc(pod); err == nil {
			shas := make([]string, len(pod.Status.ContainerStatuses))
			for _, container := range pod.Status.ContainerStatuses {
				cid := strings.Split(container.ContainerID, "//")
				shas = append(shas, cid[1])
			}
			slices.Sort(shas)

			kc.podmaps[key] = strings.Join(shas, ",")
		} else {
			kc.zlog.Error("Failed to split meta namespace cache key", zapcore.Field{Type: zapcore.ErrorType, Key: key})
		}
	}
}
func (kc *kclient) Profilling() {

	if len(kc.gsettings.Namespaces) == 0 {
		kc.profillingWithNamespace("")
	} else {
		for _, ns := range kc.gsettings.Namespaces {
			kc.profillingWithNamespace(ns)
		}
	}
}

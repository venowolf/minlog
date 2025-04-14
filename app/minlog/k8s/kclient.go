package k8s

import (
	"context"
	"os"
	"time"
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
	Start() error
}

func NewKClient(nn string, ns string, runningOnly bool) KClient {
	// creates the out-cluster config
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")

	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	if err != nil {
		log.GetLogger().Panic("failed to create in-cluster config", zap.String("Error", err.Error()))
		return nil
	}
	kc := &kclient{
		zlog: log.GetLogger(),

		clientset:   kubernetes.NewForConfigOrDie(config),
		runningOnly: runningOnly,
		namespace:   "",
		pods:        make(map[string]*v1.Pod),
		pqueue:      workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]()),
	}

	if nn != "" {
		kc.nodeName = nn
	} else {
		lo := metav1.ListOptions{
			FieldSelector: "metadata.name=" + os.Getenv("HOSTNAME"),
		}
		selfp, err := kc.clientset.CoreV1().Pods("").List(context.TODO(), lo)
		if len(selfp.Items) != 1 || err != nil {
			return nil
		}
		kc.nodeName = selfp.Items[0].Spec.NodeName
	}
	if ns != "all" && ns != "" {
		kc.namespace = ns
	}

	return kc
}

type kclient struct {
	zlog *zap.Logger

	clientset *kubernetes.Clientset

	nodeName    string
	namespace   string
	runningOnly bool

	// pods keeps all the pods which are running on this node(pods logs will be pushed to loki)
	// key: namesapace/podname
	// value: pod's object
	pods map[string]*v1.Pod

	pqueue workqueue.TypedRateLimitingInterface[string]
}

func (kc *kclient) Start() error {
	// create the pod watcher
	fs := fields.Set{"spec.nodeName": kc.nodeName}
	if kc.runningOnly {
		fs["status.phase"] = "Running"
	}

	_, informer := cache.NewInformerWithOptions(cache.InformerOptions{
		ListerWatcher: cache.NewListWatchFromClient(kc.clientset.CoreV1().RESTClient(), "pods", kc.namespace, fs.AsSelector()),
		ObjectType:    &v1.Pod{},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					kc.pqueue.Add(key)
				}
			},
			// UpdateFunc: func(old, new interface{}) {
			// 	key, err := cache.MetaNamespaceKeyFunc(new)
			// 	if err == nil {
			// 		kc.pqueue.Add(key)
			// 	}
			// },
			DeleteFunc: func(obj interface{}) {
				// IndexerInformer uses a delta queue, therefore for deletes we have to use this
				// key function.
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					kc.pqueue.Add(key)
				}
			},
		},
	})

	// Implement the logic to execute the KClient functionality

	return nil
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

	return kc.syncHandler(ctx, key)
}

func (kc *kclient) syncHandler(ctx context.Context, key string) bool {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		kc.zlog.Error("Failed to split meta namespace cache key", zapcore.Field{Type: zapcore.ErrorType, Key: key})
		return false
	}

	startTime := time.Now()
	kc.zlog.Info("Syncing pods", zap.String("namespace", namespace), zap.String("name", name), zap.Time("startTime", startTime))

	defer func() {
		kc.zlog.Info("Finished syncing pods", zap.String("namespace", namespace), zap.String("name", name), zap.Time("startTime", time.Now()))
	}()

	// Get the pod object
	pod, err := kc.clientset.CoreV1().Pods(namespace).Get(context.TODO(), name)

	if pp := kc.pods[key]; pp != nil {
		kc.zlog.Info("Pod already exists", zap.String("namespace", namespace), zap.String("name", name))
		return true
	}
	return true
}

package k8s

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
	"venowolf/minlog/app/global"
	"venowolf/minlog/app/minlog/grafanaalloy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type KClient interface {
	Run(ctx context.Context)
	Profilling(rYes bool)
}

func NewKClient(nn, ns string, inCluster, runningOnly bool) KClient {
	var config *rest.Config = nil
	var err error = nil
	if inCluster {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()

	} else {
		// creates the out-cluster config
		config, err = clientcmd.BuildConfigFromFlags("", "/root/.kube/config")

	}
	if err != nil {
		//log.GetLogger().Panic("failed to create in-cluster config", zap.String("FatalError", err.Error()))
		return nil
	}
	kc := &kclient{
		gsettings: global.NewGSettings(nn, ns),
		log:       klog.Background().WithName("kubernetes-client"),

		clientset:   kubernetes.NewForConfigOrDie(config),
		runningOnly: runningOnly,
		podmaps:     make(map[string]*grafanaalloy.PodInfo),
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
		kc.gsettings.NameSpaces = strings.Split(ns, ",")
	}

	return kc
}

type kclient struct {
	gsettings *global.GlobalSettings

	clientset *kubernetes.Clientset

	runningOnly bool

	// pods keeps all the pods which are running on this node(pods logs will be pushed to loki)
	// key: namesapace/podname
	// value: pod's object
	podmaps map[string]*grafanaalloy.PodInfo

	pqueue workqueue.TypedRateLimitingInterface[string]

	log klog.Logger
}

func (kc *kclient) initInformerWithNameSpace(nn string) cache.SharedInformer {
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
				kc.pqueue.Add("A#" + key)
				kc.log.Info("Pod added " + key)
			}
		},
		UpdateFunc: func(oobj, nobj interface{}) {
			if oobj.(*v1.Pod).ResourceVersion == nobj.(*v1.Pod).ResourceVersion {
				return // Periodic resync will send update events for all known pods.
			}
			key, err := cache.MetaNamespaceKeyFunc(nobj)
			if err == nil {
				kc.pqueue.Add("U#" + key)
				kc.log.Info("Pod updated " + key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				kc.pqueue.Add("D#" + key)
				kc.log.Info("Pod deleted " + key)
			}
		},
	})

	return podInformer
}
func (kc *kclient) runInformers(ctx context.Context, stop <-chan struct{}) {
	if len(kc.gsettings.NameSpaces) == 0 {
		informer := kc.initInformerWithNameSpace("")
		go informer.Run(stop)

		if !cache.WaitForCacheSync(stop, informer.HasSynced) {
			kc.log.Error(fmt.Errorf("Error:CacheSyncError"), "fail to start informer")
			return
		}
	} else {
		for _, ns := range kc.gsettings.NameSpaces {
			informer := kc.initInformerWithNameSpace(ns)
			go informer.Run(stop)
			if !cache.WaitForCacheSync(stop, informer.HasSynced) {
				kc.log.Error(fmt.Errorf("Error:CacheSyncError"), "fail to start informer")
				return
			}
		}
	}
}
func (kc *kclient) Run(ctx context.Context) {
	stop := make(chan struct{})
	defer func() {
		kc.pqueue.ShutDown()
		close(stop)
	}()

	// Let the workers stop when we are done
	kc.log.Info("Starting minlog controller")

	kc.runInformers(ctx, stop)

	//controller loops
	go wait.UntilWithContext(ctx, kc.startLoop, time.Second)

	<-ctx.Done()
}

func (kc *kclient) startLoop(ctx context.Context) {
	loopfunc := func(ctx context.Context) bool {
		key, quit := kc.pqueue.Get()
		if quit {
			return false
		}
		defer kc.pqueue.Done(key)

		kc.syncPodsMap(ctx, key)
		return true
	}
	for loopfunc(ctx) {
	}
}

func (kc *kclient) syncPodsMap(ctx context.Context, key string) {
	isSynced := false
	nnkey := key[2:]
	ns, name, err := cache.SplitMetaNamespaceKey(nnkey)
	// string convert to
	if err != nil {
		kc.log.Error(err, "Failed to split meta namespace cache key \""+key+"\"")
		return
	}

	defer func() {
		kc.log.Info("PodMap had already synced...")
	}()
	switch key[0:1] {
	case "A", "U":
		pod, e := kc.clientset.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if e != nil {
			kc.log.Error(e, "Failed to get pod, podname:"+name)
			return
		}

		shas := make([]string, 0, len(pod.Status.ContainerStatuses))
		for _, container := range pod.Status.ContainerStatuses {
			cid := strings.Split(container.ContainerID, "//")
			shas = append(shas, cid[1])
		}
		slices.Sort(shas)

		mpod, nexist := kc.podmaps[nnkey]
		if nexist {
			mshas := make([]string, len(mpod.ContainerMap))
			for _, mv := range mpod.ContainerMap {
				mshas = append(mshas, mv)
			}
			slices.Sort(mshas)
			if strings.Compare(strings.Join(shas, ","), strings.Join(mshas, ",")) == 0 {
				nexist = true
			}
		}
		if !nexist {
			pi := &grafanaalloy.PodInfo{
				NameSpace:    ns,
				PodName:      name,
				ContainerMap: kc.getPods(pod),
				AppLogs:      kc.gsettings.AppLogs,
				ServiceName:  getServiceNameOfPod(pod),
			}
			kc.podmaps[nnkey] = pi
			isSynced = true
		}
	case "D":
		delete(kc.podmaps, key)
		kc.log.Info("Delete " + key)
		isSynced = true
	}
	if isSynced {
		if err := grafanaalloy.NewAlloyFile(kc.gsettings).GenAlloyFile(kc.podmaps); err == nil {
			grafanaalloy.NewGAlloy().Reload()
		}
	}
}

func (kc *kclient) profillingWithNameSpace(nn string) {
	fs := fields.Set{"spec.nodeName": kc.gsettings.NodeName}
	if kc.runningOnly {
		fs["status.phase"] = "Running"
	}
	pods, err := kc.clientset.CoreV1().Pods(nn).List(context.TODO(), metav1.ListOptions{FieldSelector: fs.AsSelector().String()})
	if err != nil {
		kc.log.Error(err, "Failed to list pods, selector:"+fs.AsSelector().String())
		return
	}
	for _, pod := range pods.Items {
		shas := []string{}
		for _, container := range pod.Status.ContainerStatuses {
			if cid := strings.Split(container.ContainerID, "//"); len(cid) > 1 {
				shas = append(shas, cid[1])
			}
		}
		slices.Sort(shas)

		kc.podmaps[pod.Namespace+"/"+pod.Name] = &grafanaalloy.PodInfo{
			PodName:      pod.Name,
			NameSpace:    pod.Namespace,
			ServiceName:  getServiceNameOfPod(&pod),
			ContainerMap: kc.getPods(&pod),
			AppLogs:      kc.gsettings.AppLogs,
		}
	}
}

func (kc *kclient) Profilling(rYes bool) {
	if len(kc.gsettings.NameSpaces) == 0 {
		kc.profillingWithNameSpace("")
	} else {
		for _, ns := range kc.gsettings.NameSpaces {
			kc.profillingWithNameSpace(ns)
		}
	}
	if err := grafanaalloy.NewAlloyFile(kc.gsettings).GenAlloyFile(kc.podmaps); err == nil && rYes {
		grafanaalloy.NewGAlloy().Reload()
	}
}

func (kc *kclient) getPods(pod *v1.Pod) map[string]string {
	containers := map[string]string{}
	for _, ap := range pod.Status.ContainerStatuses {
		if cid := strings.Split(ap.ContainerID, "//"); len(cid) > 1 {
			containers[ap.Name] = cid[1]
		}
	}
	return containers
}

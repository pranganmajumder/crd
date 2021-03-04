


package main

import (
	"flag"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	appscodeclientset "github.com/pranganmajumder/crd/pkg/client/clientset/versioned"
	appscodeinformers "github.com/pranganmajumder/crd/pkg/client/informers/externalversions"


)



func main() {
	var kubeconfig string
	var master string

	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")



	cfg, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes appscodeclientset: %s", err.Error())
	}

	appscodeClient, err := appscodeclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example appscodeclientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	appscodeInformerFactory := appscodeinformers.NewSharedInformerFactory(appscodeClient, time.Second*30)

	controller := NewController(kubeClient, appscodeClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		appscodeInformerFactory.Appscode().V1alpha1().Apployments())


	// set up signals so we handle the first shutdown signal gracefully
	stopCh := make(chan struct{})
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered appscodeinformers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	appscodeInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}



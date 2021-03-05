/*
Copyright 2017 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	pranganclientset "github.com/pranganmajumder/crd/pkg/client/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"

	appscodev1alpha1 "github.com/pranganmajumder/crd/pkg/apis/appscode.com/v1alpha1"
	informers "github.com/pranganmajumder/crd/pkg/client/informers/externalversions/appscode.com/v1alpha1"
	listers "github.com/pranganmajumder/crd/pkg/client/listers/appscode.com/v1alpha1"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
)

const controllerAgentName = "sample-controller"

const (
	SuccessSynced = "Synced"
	ErrResourceExists = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	MessageResourceSynced = "Foo synced successfully"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset    kubernetes.Interface
	pranganclientset pranganclientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	apploymentsLister listers.ApploymentLister
	apploymentsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

}

// NewController returns a new sample controller
func NewController(kubeclientset kubernetes.Interface, pranganclientset pranganclientset.Interface, deploymentInformer appsinformers.DeploymentInformer,
	apploymentInformer informers.ApploymentInformer) *Controller {

	controller := &Controller{
		kubeclientset:     kubeclientset,
		pranganclientset:  pranganclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		apploymentsLister: apploymentInformer.Lister(),
		apploymentsSynced: apploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Apployments"),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	apploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueApployment,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueApployment(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Foo resource will enqueue that Foo resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Apployment controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.apploymentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting ----------------- controller or worker ")
	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	// Launch two workers to process Apployment resources
	for i := 0; i < threadiness; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	// wait until we're told to stop
	<-stopCh
	klog.Info("Shutting down ------------------ controller or worker ")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// you always have to indicate to the queue that you've completed a piece of
	// work
	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// do your work on the key.  This method will contains your "do stuff" logic
		// Apployment resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Apployment resource with this namespace/name
	apployment, err := c.apploymentsLister.Apployments(namespace).Get(name)
	if err != nil {
		// The Apployment resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("apployment '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := apployment.Spec.ApploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Apployment.spec
	deployment, err := c.deploymentsLister.Deployments(apployment.Namespace).Get(deploymentName)
	//serviceClient, err := c.kubeclientset.CoreV1().Services(apployment.Namespace)


	// If the resource doesn't exist, we'll create it

	if errors.IsNotFound(err) {
		fmt.Println("Creating Deployment .......... ")
		deployment, err = c.kubeclientset.AppsV1().Deployments(apployment.Namespace).Create(context.TODO() ,newDeployment(apployment), metav1.CreateOptions{})
		if err != nil{
			fmt.Errorf("Error %v detected while creating Deployment", err.Error())
		}

		fmt.Println("Creating Service---------------")
		service, err := c.kubeclientset.CoreV1().Services(apployment.Namespace).Create(context.TODO(), newService(apployment), metav1.CreateOptions{})

		if err != nil{
			fmt.Errorf("error %v detected while creating Service %q", err.Error(), service.GetObjectMeta().GetName())
		}
		fmt.Println("service Created Successfully \n")
	}


	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this Apployment resource, we should log
	// a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(deployment, apployment) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		//c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg) // to run the program we omitted it
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the Apployment resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if apployment.Spec.Replicas != nil && *apployment.Spec.Replicas != *deployment.Spec.Replicas {
		klog.V(4).Infof("Apployment %s replicas: %d, deployment replicas: %d", name, *apployment.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1().Deployments(apployment.Namespace).Update( context.TODO(), newDeployment(apployment), metav1.UpdateOptions{})
		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
	}



	// Finally, we update the status block of the Apployment resource to reflect the
	// current state of the world
	err = c.updateApploymentStatus(apployment, deployment)
	if err != nil {
		return err
	}

	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateApploymentStatus(apployment *appscodev1alpha1.Apployment, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	apploymentCopy := apployment.DeepCopy()
	apploymentCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Apployment resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.pranganclientset.AppscodeV1alpha1().Apployments(apployment.Namespace).UpdateStatus(context.TODO(),  apploymentCopy, metav1.UpdateOptions{})
	return err
}

// enqueueApployment takes a Apployment resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Apployment.
func (c *Controller) enqueueApployment(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Apployment" {
			return
		}

		foo, err := c.apploymentsLister.Apployments(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueApployment(foo)
		return
	}
}



// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(apployment *appscodev1alpha1.Apployment) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "Appscode",
		"controller": apployment.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apployment.Spec.ApploymentName,
			Namespace: apployment.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apployment, appscodev1alpha1.SchemeGroupVersion.WithKind("Apployment")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: apployment.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  apployment.Name,
							Image: apployment.Spec.Image,
						},
					},
				},
			},
		},
	}
}

func newService(apployment *appscodev1alpha1.Apployment) *corev1.Service  {

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: 			apployment.ObjectMeta.Name,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apployment, appscodev1alpha1.SchemeGroupVersion.WithKind("Apployment")),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: apployment.Spec.ContainerPort,
					NodePort: apployment.Spec.NodePort, //not mendatory
				},
			},
			//Selector: apployment.Spec.Label,
			Type: corev1.ServiceType(apployment.Spec.ServiceType),
		},
	}
}
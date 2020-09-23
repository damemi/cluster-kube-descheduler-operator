package operator

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/operatorclient"
	"reflect"
	"strconv"
	"time"

	deschedulerv1 "github.com/openshift/api/operator/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	operatorconfigclientv1 "github.com/openshift/client-go/operator/clientset/versioned/typed/operator/v1"
	operatorclientinformers "github.com/openshift/client-go/operator/informers/externalversions/operator/v1"
	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/v410_00_assets"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const DefaultImage = "quay.io/openshift/origin-descheduler:latest"

// deschedulerCommand provides descheduler command with policyconfigfile mounted as volume and log-level for backwards
// compatibility with 3.11
var DeschedulerCommand = []string{"/bin/descheduler", "--policy-config-file", "/policy-dir/policy.yaml", "--v", "2"}

type TargetConfigReconciler struct {
	ctx               context.Context
	deschedulerClient operatorconfigclientv1.KubeDeschedulerInterface
	operatorClient    v1helpers.OperatorClient
	kubeClient        kubernetes.Interface
	eventRecorder     events.Recorder
	queue             workqueue.RateLimitingInterface
}

func NewTargetConfigReconciler(
	ctx context.Context,
	operatorConfigClient operatorconfigclientv1.KubeDeschedulerInterface,
	operatorClientInformer operatorclientinformers.KubeDeschedulerInformer,
	operatorClient v1helpers.OperatorClient,
	kubeClient kubernetes.Interface,
	eventRecorder events.Recorder,
) *TargetConfigReconciler {
	c := &TargetConfigReconciler{
		ctx:               ctx,
		deschedulerClient: operatorConfigClient,
		operatorClient:    operatorClient,
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		queue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TargetConfigReconciler"),
	}

	operatorClientInformer.Informer().AddEventHandler(c.eventHandler())

	return c
}

func (c TargetConfigReconciler) sync() error {
	descheduler, err := c.deschedulerClient.Get(c.ctx, operatorclient.OperatorConfigName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "unable to get operator configuration", "namespace", operatorclient.OperatorNamespace, "kubedescheduler", operatorclient.OperatorConfigName)
		return err
	}

	if descheduler.Spec.DeschedulingIntervalSeconds == nil {
		return fmt.Errorf("descheduler should have an interval set")
	}

	deployment, _, err := c.manageDeployment(descheduler, false)
	if err != nil {
		return err
	}

	_, _, err = v1helpers.UpdateStatus(c.operatorClient, func(status *operatorv1.OperatorStatus) error {
		resourcemerge.SetDeploymentGeneration(&status.Generations, deployment)
		return nil
	})
	return err
}

func (c *TargetConfigReconciler) manageDeployment(descheduler *deschedulerv1.KubeDescheduler, forceDeployment bool) (*appsv1.Deployment, bool, error) {
	required := resourceread.ReadDeploymentV1OrDie(v410_00_assets.MustAsset("v4.1.0/kube-descheduler/deployment.yaml"))
	required.Name = descheduler.Name
	required.Namespace = descheduler.Namespace
	required.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "v1beta1",
			Kind:       "KubeDescheduler",
			Name:       descheduler.Name,
			UID:        descheduler.UID,
		},
	}
	required.Spec.Template.Spec.Containers[0].Image = descheduler.Spec.Image
	required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args,
		fmt.Sprintf("--descheduling-interval=%ss", strconv.Itoa(int(*descheduler.Spec.DeschedulingIntervalSeconds))))

	// Set descheduler deployment to read policy configmap from descheduler.Spec.Policy field
	required.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name = descheduler.Spec.Policy

	// Add any additional flags that were specified
	if len(descheduler.Spec.Flags) > 0 {
		required.Spec.Template.Spec.Containers[0].Args = append(required.Spec.Template.Spec.Containers[0].Args, descheduler.Spec.Flags...)
	}

	if !forceDeployment {
		existingDeployment, err := c.kubeClient.AppsV1().Deployments(required.Namespace).Get(c.ctx, descheduler.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				forceDeployment = true
			} else {
				return nil, false, err
			}
		} else {
			forceDeployment = deploymentChanged(existingDeployment, required)
		}
	}
	// FIXME: this method will disappear in 4.6 so we need to fix this ASAP
	return resourceapply.ApplyDeploymentWithForce(
		c.kubeClient.AppsV1(),
		c.eventRecorder,
		required,
		resourcemerge.ExpectedDeploymentGeneration(required, descheduler.Status.Generations),
		forceDeployment)
}

func deploymentChanged(existing, new *appsv1.Deployment) bool {
	newArgs := sets.NewString(new.Spec.Template.Spec.Containers[0].Args...)
	existingArgs := sets.NewString(existing.Spec.Template.Spec.Containers[0].Args...)
	return existing.Name != new.Name ||
		existing.Namespace != new.Namespace ||
		existing.Spec.Template.Spec.Containers[0].Image != new.Spec.Template.Spec.Containers[0].Image ||
		existing.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name != new.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name ||
		!reflect.DeepEqual(newArgs, existingArgs)
}

// Run starts the kube-scheduler and blocks until stopCh is closed.
func (c *TargetConfigReconciler) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Infof("Starting TargetConfigReconciler")
	defer klog.Infof("Shutting down TargetConfigReconciler")

	// doesn't matter what workers say, only start one.
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *TargetConfigReconciler) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *TargetConfigReconciler) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.sync()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
func (c *TargetConfigReconciler) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}

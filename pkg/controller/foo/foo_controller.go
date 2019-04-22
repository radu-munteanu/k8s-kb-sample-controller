/*
Copyright 2019 Radu Munteanu.

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

package foo

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"syscall"
	"time"

	toolsv1beta1 "gitlab.com/radu-munteanu/k8s-kb-sample-controller/pkg/apis/tools/v1beta1"
	"gitlab.com/radu-munteanu/k8s-kb-sample-controller/pkg/controller/foo/leaderelectioninfo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	k8scorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Foo Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	lei, err := initLeaderElection(kubernetes.NewForConfigOrDie(mgr.GetConfig()).CoreV1())
	if err != nil {
		log.Info("Could not init leader election: " + err.Error())
		syscall.Exit(200)
	}

	return &ReconcileFoo{Client: mgr.GetClient(), scheme: mgr.GetScheme(), lei: lei}
}

func initLeaderElection(coreV1 k8scorev1.CoreV1Interface) (*leaderelectioninfo.LeaderElectionInfo, error) {
	leNamespace := "default" //"leaderelection"

	identity := fmt.Sprintf("foocontrollerelection-id-%d", rand.Int())
	log.Info("I'm " + identity)
	resourceLockConfig := resourcelock.ResourceLockConfig{
		Identity:      identity,
		EventRecorder: &record.FakeRecorder{},
	}
	lock, err := resourcelock.New(resourcelock.EndpointsResourceLock, leNamespace, "foocontrollerelection",
		coreV1, resourceLockConfig)

	if err != nil {
		return nil, fmt.Errorf("Could not make Resource Lock: %s", err)
	}

	retryPeriod := 2 * time.Second
	renewDeadline := 3 * retryPeriod
	leaseDuration := 3 * renewDeadline

	lei := leaderelectioninfo.New(identity)

	lec := leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				log.Info("Leader Elector: leading...")
			},
			// OnStoppedLeading is called when a LeaderElector client stops leading
			OnStoppedLeading: func() {
				log.Info("Leader Elector: stopped leading.")
			},
			// OnNewLeader is called when the client observes a leader that is
			// not the previously observed leader. This includes the first observed
			// leader when the client starts.
			OnNewLeader: func(l string) {
				log.Info("Leader Elector: new leader: " + l)
				lei.SetLeader(l)
			},
		},
	}
	le, err := leaderelection.NewLeaderElector(lec)
	if err != nil {
		return nil, fmt.Errorf("Could not make the Leader Elector: %s", err)
	}

	go func() {
		le.Run(context.Background())
		// "Run starts the leader election loop"
		// - "acquire loops calling tryAcquireOrRenew and returns immediately when tryAcquireOrRenew succeeds"
		// - "renew loops calling tryAcquireOrRenew and returns immediately when tryAcquireOrRenew fails"
		// - "tryAcquireOrRenew tries to acquire a leader lease if it is not already acquired,
		// else it tries to renew the lease if it has already been acquired. Returns true
		// on success else returns false"

		// if renew fails, probably something is wrong with the cluster's resources, so we can kill this
		// controller to spawn a fresh one
		syscall.Exit(200)
	}()

	return lei, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("foo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Foo
	err = c.Watch(&source.Kind{Type: &toolsv1beta1.Foo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Foo - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &toolsv1beta1.Foo{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileFoo{}

// ReconcileFoo reconciles a Foo object
type ReconcileFoo struct {
	client.Client
	scheme *runtime.Scheme
	lei    *leaderelectioninfo.LeaderElectionInfo
}

// Reconcile reads that state of the cluster for a Foo object and makes changes based on the state read
// and what is in the Foo.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tools.example.com,resources=foos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tools.example.com,resources=foos/status,verbs=get;update;patch
func (r *ReconcileFoo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Request received:" + request.String())

	for r.lei.GetLeader() == "" {
		time.Sleep(1 * time.Second)
	}

	if r.lei.IsLeader() {
		log.Info("I am the leader! Going to work.")
	} else {
		log.Info("I am not the leader. Leader is: " + r.lei.GetLeader() + ". Doing nothing.")
		return reconcile.Result{}, nil
	}

	// Fetch the Foo instance
	instance := &toolsv1beta1.Foo{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	specReplicas := instance.Spec.Replicas

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.DeploymentName,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Spec.DeploymentName},
			},
			Replicas: &specReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Spec.DeploymentName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// updating the status
	instance.Status.AvailableReplicas = found.Status.AvailableReplicas
	err = r.Status().Update(context.Background(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

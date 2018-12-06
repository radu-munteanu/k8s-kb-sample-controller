# Kubernetes KubeBuilder Sample Controller

## About
This is the [sample-controller](https://github.com/kubernetes/sample-controller) redone using the [KubeBuilder](https://github.com/kubernetes-sigs/kubebuilder). It also contains information and code for deploying the controller manager to Kubernetes.

> The custom controller is actually a custom controller manager. A manager is a process that can encapsulate multiple controllers. It just happens that, in our case, there is only one controller instance.

**Prerequisites:**
- MiniKube
- KubeCtl
- GoLang
- Docker 

Last versions tested:
- MiniKube: 0.30
- Kubernetes: 1.12 (`minikube start --kubernetes-version v1.12.0`)
- KubeCtl: 1.12
- KubeBuilder: 1.0.5
- GoLang: 1.11.3
- Docker: ce-18.06.1

## Table of Contents

* [Build This Project](#build-this-project)
* [Make a New Controller from Scratch](#make-a-new-controller-from-scratch)
    * [Code, Build and Run The Controller Locally](#code-build-and-run-the-controller-locally)
* [Deploy Controller to Kubernetes](#deploy-controller-to-kubernetes)
    * [Create The Docker Image and Push It to a Docker Registry](#create-the-docker-image-and-push-it-to-a-docker-registry)
    * [Apply Additional Resources](#apply-additional-resources)
    * [Create The Controller Resource and Apply It to The K8s Cluster](#create-the-controller-resource-and-apply-it-to-the-k8s-cluster)
* [Fixes](#fixes)
* [What's Next](#whats-next)

## Build This Project
To build the controller without starting from scratch, you can clone this repo into your Go sources directory.

Build
```bash
cd "${GOPATH}/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller"

make
```

You can now jump to [Deploy Controller to Kubernetes](#deploy-controller-to-kubernetes)

If you want to know more about the code, please read all the chapters below.

## Make a New Controller from Scratch
The following steps will guide you through replicating the whole project without cloning any sources from this repo. Use this to learn how to make a custom controller with a custom resource.

### Code, Build and Run The Controller Locally
Install KubeBuilder and project dir
```bash
cd ~
# get latest kubebuilder
kb_version="1.0.5"
curl -o "kubebuilder_${kb_version}_linux_amd64.tar.gz" -L "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${kb_version}/kubebuilder_${kb_version}_linux_amd64.tar.gz"
rm -rf kubebuilder_${kb_version}_linux_amd64 && tar -zxvf "kubebuilder_${kb_version}_linux_amd64.tar.gz"
sudo rm -rf '/usr/local/kubebuilder' && sudo mv "kubebuilder_${kb_version}_linux_amd64" '/usr/local/kubebuilder'

# one time only
sudo bash -c 'cat > /etc/profile.d/kubebuilder.sh <<EOF
export PATH=\$PATH:/usr/local/kubebuilder/bin
EOF'

# create you project directory structure. Example:
mkdir -p "${GOPATH}/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller"

# check version in a go project
cd "${GOPATH}/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller"
kubebuilder version
```

Initial Boilerplating (Generating Code for Controller and Foo Custom Resource)
```bash
# install go dep if not installer
go get -u github.com/golang/dep/cmd/dep
sudo mkdir -p /usr/local/bin
sudo cp $GOPATH/bin/dep /usr/local/bin

# generate controller boilerplate
kubebuilder init --dep --domain example.com --license apache2 --owner "Radu Munteanu"

# generate resource boilerplate
kubebuilder create api --resource --controller --group tools --version v1beta1 --kind Foo

# test build - we are going to use the built in make in the future 
go install gitlab.com/radu-munteanu/k8s-kb-sample-controller/cmd/manager && printf 'OK!\n'

# cleanup binary generated
rm -f "${GOPATH}/bin/manager"
```

Add Spec and Status variables to Foo resource (`pkg/apis/tools/v1beta1/foo_types.go`)
```go
...

// FooSpec defines the desired state of Foo
type FooSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeploymentName string `json:"deploymentName"`
	Replicas       int32  `json:"replicas"`
}

// FooStatus defines the observed state of Foo
type FooStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is the Schema for the foos API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec,omitempty"`
	Status FooStatus `json:"status,omitempty"`
}

...
```

> Note that Status is a [subresource](https://book.kubebuilder.io/basics/status_subresource.html): `// +kubebuilder:subresource:status`

Generate Boilerplate for updated resource
```bash
make
```

Add Status update code in Reconcile func (`pkg/controller/foo/foo_controller.go`)
```go

func (r *ReconcileFoo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
...

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
		log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
```

Build again
```bash
make
```

Start the Controller Manager.
```
# start K8s
minikube start --kubernetes-version v1.12.0

# create the CRD
# - option 1
kubectl apply -f config/crds
# - option 2
make install

bin/manager -kubeconfig=$HOME/.kube/config
```

In another terminal, create a resource and check what happens.
```bash
cd "${GOPATH}/src/gitlab.com/radu-munteanu/k8s-kb-sample-controller"
cat <<EOF > 'config/samples/foo-sample-01.yaml'
apiVersion: tools.example.com/v1beta1
kind: Foo
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: foo-sample-01
spec:
  # Add fields here
  deploymentName: dep-foo-sample-01
  replicas: 1
EOF

kubectl apply -f config/samples/foo-sample-01.yaml
foo.tools.example.com/foo-sample-01 created


kubectl get Foo foo-sample-01

# wait a few seconds for the deployment's pod to start

kubectl describe Foo foo-sample-01

kubectl get deployment foo-sample-01-deployment

kubectl describe deployment foo-sample-01-deployment

# (optional) you can check the Foo resource also by using the sample plugin (https://gitlab.com/radu-munteanu/k8s-sample-plugin)
kubectl sample
```

## Deploy Controller to Kubernetes
Close the running controller manager (`CTRL + C`).

### Create The Docker Image and Push It to a Docker Registry

Build the docker image.
```bash
make docker-build
```

Check the docker images.
```bash
docker images
REPOSITORY                          TAG                 IMAGE ID            CREATED             SIZE
controller                          latest              0f99cb7477e8        10 seconds ago      122MB
...
```

Make a Docker Registry on MiniKube with a shared volume from your machine user's home
```bash
minikube ssh "docker run -d -p 5000:5000 --restart=always --name registry -v /hosthome/$(whoami)/registry/:/var/lib/registry registry"
```

For the MiniKube VM, port forward 5000 to 5000 on the Network Adapter 1 (NAT).

Tag the image.
```bash
docker tag <image_id> localhost:5000/kb-manager:latest
## or, presuming the image is the first ...
docker tag $(docker images -q | head -1) localhost:5000/kb-manager:latest
```

Push the local image to the Docker Registry on MiniKube.
```bash
docker push localhost:5000/kb-manager:latest
```

Check the Docker Registry
```bash
minikube ssh "curl -X GET http://localhost:5000/v2/_catalog"
```

### Apply Additional Resources
```bash
cat 'config/rbac/rbac_role.yaml' > 'config/rbac/kb-rbac_role.yaml'

cat <<EOF >> 'config/rbac/kb-rbac_role.yaml'
- apiGroups:
  - tools.example.com
  resources:
  - foos/status
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
EOF

kubectl apply -f config/rbac/kb-rbac_role.yaml

kubectl apply -f config/rbac/rbac_role_binding.yaml
```

### Create The Controller Resource and Apply It to The K8s Cluster
Make use of the existing K8s config found at `config/manager/manager.yaml`:
```bash
sed 's/image: controller:latest/image: localhost:5000\/kb-manager:latest/g' 'config/manager/manager.yaml' > 'config/manager/kb-manager.yaml'
```

Apply the config.
```bash
kubectl apply -f 'config/manager/kb-manager.yaml'
```

Create a resource and check what happens.
```bash
cat <<EOF > 'config/samples/foo-sample-02.yaml'
apiVersion: tools.example.com/v1beta1
kind: Foo
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: foo-sample-02
spec:
  # Add fields here
  deploymentName: dep-foo-sample-02
  replicas: 2
EOF

kubectl apply -f config/samples/foo-sample-02.yaml

kubectl get Foo foo-sample-02

# wait a few seconds for the deployment's pod to start

kubectl describe Foo foo-sample-02

kubectl get deployment foo-sample-02-deployment

kubectl describe deployment foo-sample-02-deployment

# we can see one replica in the deployment instead of two. why?
```

## Fixes
For foo-sample-02, we're not seeing two replicas, but just one. That's because the current code defines a deployment without the given replicas number in the spec.

Update the Reconcile function (`pkg/controller/foo/foo_controller.go`):
```go
func (r *ReconcileFoo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
...

	specReplicas := instance.Spec.Replicas

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Replicas: &specReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
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

...
```

Rebuild and deploy
```
make

make docker-build

docker images
REPOSITORY                          TAG                 IMAGE ID            CREATED             SIZE
controller                          latest              f53d08bd7283        2 seconds ago       122MB

docker tag $(docker images -q | head -1) localhost:5000/kb-manager:latest

docker push localhost:5000/kb-manager:latest

kubectl delete statefulset controller-manager -n system

kubectl apply -f 'config/manager/kb-manager.yaml'
```

Check the foo-sample-02 after a few moments
```bash
kubectl describe Foo foo-sample-02

kubectl describe deployment foo-sample-02-deployment

# (optional) you can check the Foo resource also by using the sample plugin (https://gitlab.com/radu-munteanu/k8s-sample-plugin)
kubectl sample
```

## What's Next

What to look at next:

- Finalizers - for special cleanup logic // "Created objects are automatically garbage collected" when the resource is not found
    - [https://book.kubebuilder.io/beyond_basics/using_finalizers.html](https://book.kubebuilder.io/beyond_basics/using_finalizers.html)

- Leader Election
    - [https://github.com/kubernetes-sigs/kubebuilder/issues/230](https://github.com/kubernetes-sigs/kubebuilder/issues/230)
    - [https://github.com/kubernetes-sigs/kubebuilder/projects/2#card-10213326](https://github.com/kubernetes-sigs/kubebuilder/projects/2#card-10213326)
    - there is some leader election code in the controller manager (in sigs.k8s.io)

- Events
    - [https://book.kubebuilder.io/beyond_basics/creating_events.html](https://book.kubebuilder.io/beyond_basics/creating_events.html)

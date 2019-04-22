# Kubernetes KubeBuilder Sample Controller

## About
This is the [sample-controller](https://github.com/kubernetes/sample-controller) redone using the [KubeBuilder](https://github.com/kubernetes-sigs/kubebuilder). It also contains information and code for deploying the controller manager to Kubernetes.

> The custom controller is actually a custom controller manager. A manager is a process that can encapsulate multiple controllers. It just happens that, in our case, there is only one controller instance.

**Prerequisites:**
- Virtualbox
- MiniKube
- Kubernetes
- Docker
- KubeCtl

Last versions tested:
- Fedora 29 (kernel 4.20.3-200 - updated at 2019-01-23)
- VirtualBox 6.0.2
- MiniKube: 0.33.1
- Kubernetes: 1.13 (`minikube start --kubernetes-version v1.13.0`)
- Docker: 18.06.1-ce
- KubeCtl: 1.13.2

**Go and KubeBuilder versions tested:**
- GoLang: 1.11.4
- KubeBuilder: 1.0.7

## Table of Contents

* [Build This Project](#build-this-project)
* [Install Prerequisites](#install-prerequisites)
    * [VirtualBox 6.0.2](#virtualbox-602)
    * [MiniKube 0.33.1](#minikube-0331)
    * [Kubernetes 1.13](#kubernetes-113)
    * [Docker 18.06.1](#docker-18061)
    * [Kubectl 1.13.0](#kubectl-1130)
* [Make a New Controller from Scratch](#make-a-new-controller-from-scratch)
    * [Code, Build and Run The Controller Locally](#code-build-and-run-the-controller-locally)
* [Deploy Controller to Kubernetes](#deploy-controller-to-kubernetes)
    * [Create The Docker Image and Push It to a Docker Registry](#create-the-docker-image-and-push-it-to-a-docker-registry)
    * [Apply Additional Resources](#apply-additional-resources)
    * [Create The Controller Resource and Apply It to The K8s Cluster](#create-the-controller-resource-and-apply-it-to-the-k8s-cluster)
* [Known Issues](#known-issues)
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

## Install Prerequisites
### VirtualBox 6.0.2
[https://www.virtualbox.org/wiki/Linux_Downloads](https://www.virtualbox.org/wiki/Linux_Downloads)

```bash
curl -o "VirtualBox-6.0-6.0.2_128162_fedora29-1.x86_64.rpm" -L  https://download.virtualbox.org/virtualbox/6.0.2/VirtualBox-6.0-6.0.2_128162_fedora29-1.x86_64.rpm
curl -o "oracle_vbox.asc" -L https://www.virtualbox.org/download/oracle_vbox.asc
```

```bash
gpg2 --show-keys --fingerprint --keyid-format SHORT oracle_vbox.asc
pub   dsa1024/98AB5139 2010-05-18 [SC]
      Key fingerprint = 7B0F AB3A 13B9 0743 5925  D9C9 5442 2A4B 98AB 5139
uid                    Oracle Corporation (VirtualBox archive signing key) <info@virtualbox.org>
sub   elg2048/281DDC4B 2010-05-18 [E]
```
> From [https://www.virtualbox.org/wiki/Linux_Downloads](https://www.virtualbox.org/wiki/Linux_Downloads):
>
> The key fingerprint is
>
> 7B0F AB3A 13B9 0743 5925  D9C9 5442 2A4B 98AB 5139
> Oracle Corporation (VirtualBox archive signing key) <info@virtualbox.org>

```bash
sudo rpm --import oracle_vbox.asc

rpm --checksig VirtualBox-6.0-6.0.2_128162_fedora29-1.x86_64.rpm
VirtualBox-6.0-6.0.2_128162_fedora29-1.x86_64.rpm: digests signatures OK
```

```bash
KERNEL_VERSION=$(uname -r)

sudo dnf install "kernel-devel-${KERNEL_VERSION}"

sudo dnf install VirtualBox-6.0-6.0.2_128162_fedora29-1.x86_64.rpm
```

### MiniKube 0.33.1
[https://github.com/kubernetes/minikube/releases/tag/v0.33.1](https://github.com/kubernetes/minikube/releases/tag/v0.33.1)

```bash
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.33.1/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
```

### Kubernetes 1.13
```bash
minikube start --kubernetes-version v1.13.0
```

### Docker 18.06.1
Check Docker version from Minikube
```bash
minikube ssh "docker version"
========================================
kubectl could not be found on your path. kubectl is a requirement for using minikube
To install kubectl, please run the following:

curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.13.2/bin/linux/amd64/kubectl && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl

To disable this message, run the following:

minikube config set WantKubectlDownloadMsg false
========================================
Client:
 Version:           18.06.1-ce
 API version:       1.38
 Go version:        go1.10.3
 Git commit:        e68fc7a
 Built:             Tue Aug 21 17:20:43 2018
 OS/Arch:           linux/amd64
 Experimental:      false

Server:
 Engine:
  Version:          18.06.1-ce
  API version:      1.38 (minimum version 1.12)
  Go version:       go1.10.3
  Git commit:       e68fc7a
  Built:            Tue Aug 21 17:28:38 2018
  OS/Arch:          linux/amd64
  Experimental:     false
```

Install docker from [https://download.docker.com/linux/centos/7/x86_64/stable/Packages/](https://download.docker.com/linux/centos/7/x86_64/stable/Packages/)
```bash
curl -Lo docker-ce-18.06.1.ce-3.el7.x86_64.rpm https://download.docker.com/linux/centos/7/x86_64/stable/Packages/docker-ce-18.06.1.ce-3.el7.x86_64.rpm

sudo dnf install docker-ce-18.06.1.ce-3.el7.x86_64.rpm
```

Test docker
```bash
sudo systemctl enable docker && sudo systemctl start docker

docker version

sudo docker version

# add the current user to docker group 
sudo usermod -a -G docker <user>

# log out and log in with the user

# restart minikube
minikube start
```

### Kubectl 1.13.0
(from minikube output, with version changed to match Kubernetes release)

```bash
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.13.0/bin/linux/amd64/kubectl && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl

kubectl version
Client Version: version.Info{Major:"1", Minor:"13", GitVersion:"v1.13.0", GitCommit:"ddf47ac13c1a9483ea035a79cd7c10005ff21a6d", GitTreeState:"clean", BuildDate:"2018-12-03T21:04:45Z", GoVersion:"go1.11.2", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"13", GitVersion:"v1.13.0", GitCommit:"ddf47ac13c1a9483ea035a79cd7c10005ff21a6d", GitTreeState:"clean", BuildDate:"2018-12-03T20:56:12Z", GoVersion:"go1.11.2", Compiler:"gc", Platform:"linux/amd64"}

kubectl get nodes
NAME       STATUS   ROLES    AGE   VERSION
minikube   Ready    master   32m   v1.13.0
```

## Make a New Controller from Scratch
The following steps will guide you through replicating the whole project without cloning any sources from this repo. Use this to learn how to make a custom controller with a custom resource.

### Code, Build and Run The Controller Locally
[https://github.com/kubernetes-sigs/kubebuilder/releases](https://github.com/kubernetes-sigs/kubebuilder/releases)

Install KubeBuilder and project dir
```bash
cd ~
# get latest kubebuilder
kb_version="1.0.7"
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

# restart console in order to reload profile or do bash -

# test kubebuilder
kubebuilder version
Version: version.Version{KubeBuilderVersion:"1.0.7", KubernetesVendor:"1.11", GitCommit:"63bd3604767ddb5042fe76b67d097840a7a282c2", BuildDate:"2018-12-20T18:41:54Z", GoOs:"unknown", GoArch:"unknown"}
```

Initial Boilerplating (Generating Code for Controller and Foo Custom Resource)
```bash
# install go dep if not installed
go get -u github.com/golang/dep/cmd/dep
sudo mkdir -p /usr/local/bin
sudo cp $GOPATH/bin/dep /usr/local/bin

# generate controller boilerplate - this may take a few minutes
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

Add Replicas and Status update code in Reconcile func (`pkg/controller/foo/foo_controller.go`)
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
...
	// updating the status
	instance.Status.AvailableReplicas = found.Status.AvailableReplicas
	err = r.Status().Update(context.Background(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
}
```

Change deployment definition so that the name given is the one in spec's DeploymentName (`pkg/controller/foo/foo_controller.go`)
```go
func (r *ReconcileFoo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
...
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
```

Build again
```bash
make
```

Start the Controller Manager.
```
# start K8s - if not started
minikube start --kubernetes-version v1.13.0

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

For the MiniKube VM, port forward 5000 to 5000 on the Network Adapter 1 (NAT). This can be done from the Settings of the minikube machine in VirtualBox app, or through a `vboxmanage` command.
```bash
vboxmanage controlvm "minikube" natpf1 "docker-registry,tcp,,5000,,5000"
```

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
kubectl apply -f config/rbac/rbac_role.yaml

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

# there should be two replicas in the deployment
```

## Known Issues
As of KubeBuilder 1.0.7 (vs 1.0.5), there are a few issues that pop-up in the controller manager logs (there are shown if you start it locally), even though the resources seem to be created/updated successfully:

- when applying a new resource:

    ```{
      "level": "error",
      "ts": 1548256859.920939,
      "logger": "kubebuilder.controller",
      "msg": "Reconciler error",
      "controller": "foo-controller",
      "request": "default/foo-sample-01",
      "error": "resource name may not be empty",
    ```

- when the resource is updated (increasing the number of replicas):

    ```{
      "level": "error",
      "ts": 1548259361.6205256,
      "logger": "kubebuilder.controller",
      "msg": "Reconciler error",
      "controller": "foo-controller",
      "request": "default/foo-sample-01",
      "error": "Operation cannot be fulfilled on deployments.apps \"foo-sample-01-deployment\": the object has been modified; please apply your changes to the latest version and try again",
    ```

## Leader Election Example
Check the last commit for Leader Election code.

## What's Next

What to look at next:

- Auth Proxy and the Metrics Service
    - look for `config/rbac/auth_proxy_` and `config/default/manager_` files  

- Finalizers - for special cleanup logic // "Created objects are automatically garbage collected" when the resource is not found
    - [https://book.kubebuilder.io/beyond_basics/using_finalizers.html](https://book.kubebuilder.io/beyond_basics/using_finalizers.html)

- Leader Election
    - [https://github.com/kubernetes-sigs/kubebuilder/issues/230](https://github.com/kubernetes-sigs/kubebuilder/issues/230)
    - [https://github.com/kubernetes-sigs/kubebuilder/projects/2#card-10213326](https://github.com/kubernetes-sigs/kubebuilder/projects/2#card-10213326)
    - there is some leader election code in the controller manager (in sigs.k8s.io)
    - check the last commit for Leader Election code

- Events
    - [https://book.kubebuilder.io/beyond_basics/creating_events.html](https://book.kubebuilder.io/beyond_basics/creating_events.html)

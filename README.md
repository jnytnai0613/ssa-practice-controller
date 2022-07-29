# ssa-practice-controller
## Description
Implement a Controller that does Server-Side Apply(SSA). Rewrite the Field of the deployment resource according to the .spec value of the CR set by the user.

SSA creates patches with applyconfiguration, which was implemented when Kuberenetes became GA in 1.22. In this case, we created a controller that rewrites the .spec definition of the deployment resource with CR. We use the function With\<field name\> to rewrite each field as follows.
```golang
deploymentApplyConfig := appsv1apply.Deployment("ssapractice-nginx", "ssa-practice-controller-system").
	WithSpec(appsv1apply.DeploymentSpec().
		WithSelector(metav1apply.LabelSelector().
			WithMatchLabels(labels)))

if ssapractice.Spec.DepSpec.Replicas != nil {
	replicas := *ssapractice.Spec.DepSpec.Replicas
	deploymentApplyConfig.Spec.WithReplicas(replicas)
}

if ssapractice.Spec.DepSpec.Strategy != nil {
	types := *ssapractice.Spec.DepSpec.Strategy.Type
	rollingUpdate := ssapractice.Spec.DepSpec.Strategy.RollingUpdate
	deploymentApplyConfig.Spec.WithStrategy(appsv1apply.DeploymentStrategy().
		WithType(types).
		WithRollingUpdate(rollingUpdate))
}

if ssapractice.Spec.DepSpec.Template == nil {
	return ctrl.Result{}, fmt.Errorf("Error: %s", "The name or image field is required in the '.Spec.DepSpec.Template.Spec.Containers[]'.")
}

podTemplate = ssapractice.Spec.DepSpec.Template
podTemplate.WithLabels(labels)
for i, v := range podTemplate.Spec.Containers {
	if v.Image == nil {
		var (
			image  string  = "nginx"
			pimage *string = &image
		)
		podTemplate.Spec.Containers[i].Image = pimage
	}
	if v.Name == nil {
		var (
			s             = strings.Split(*v.Image, ":")
			pname *string = &s[0]
		)
		podTemplate.Spec.Containers[i].Name = pname
	}
}
deploymentApplyConfig.Spec.WithTemplate(podTemplate)

owner, err := createOwnerReferences(ssapractice, r.Scheme, log)
if err != nil {
	log.Error(err, "Unable create OwnerReference")
	return ctrl.Result{}, err
}
deploymentApplyConfig.WithOwnerReferences(owner)
```

Define Custom Resource (CR) as follows, and include the .spec field of the deployment resource you actually want to rewrite.
```yaml
apiVersion: ssapractice.jnytnai0613.github.io/v1
kind: SSAPractice
metadata:
  name: ssapractice-sample
  namespace: ssa-practice-controller-system
spec:
  depSpec:
    replicas: 5
    strategy:
      type: RollingUpdate
      rollingUpdate:
        maxSurge: 30%
        maxUnavailable: 30%
    template:
      spec:
        containers:
          - name: nginx
            image: nginx:latest
```

## Description to each field of CR
The CR yaml file is located in the config/samples directory.

### .spec.depspec
| Name       | Type               | Required      |
| ---------- | ------------------ | ------------- |
| replicas   | int32              | false         |
| strategy   | DeploymentStrategy | false         |

Other fields cannot be specified.　  
Check the following reference for a description of the strategy field.  
https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/#DeploymentSpec


### .spec.depspec.template.spec.conatiners
| Name       | Type               | Required      |
| ---------- | ------------------ | ------------- |
| name       | string             | true          |
| image      | string             | true          |

The other fields are options.See the following reference for possible fields.  
https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#PodSpec

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.  
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/ssa-practice-controller:tag
```
	
2. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/ssa-practice-controller:tag
```

3. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

4. See our deployment resources

```sh
kubectl -n ssa-practice-controller-system get deployment,pod
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

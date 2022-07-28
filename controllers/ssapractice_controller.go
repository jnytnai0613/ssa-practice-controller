/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	ssapracticev1 "github.com/jnytnai0613/ssa-practice-controller/api/v1"
)

// SSAPracticeReconciler reconciles a SSAPractice object
type SSAPracticeReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var kclientset *kubernetes.Clientset

func init() {
	kclientset = getClient()
}

func getClient() *kubernetes.Clientset {
	kubeconfig := ctrl.GetConfigOrDie()
	kclientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	return kclientset
}

//+kubebuilder:rbac:groups=ssapractice.jnytnai0613.github.io,resources=ssapractices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ssapractice.jnytnai0613.github.io,resources=ssapractices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ssapractice.jnytnai0613.github.io,resources=ssapractices/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func createOwnerReferences(ssapractice ssapracticev1.SSAPractice, scheme *runtime.Scheme, log logr.Logger) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(&ssapractice, scheme)
	if err != nil {
		log.Error(err, "Unable get GVK")
		return nil, err
	}

	owner := metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(ssapractice.GetName()).
		WithUID(ssapractice.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)

	return owner, nil
}

func (r *SSAPracticeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		deploymentClient = kclientset.AppsV1().Deployments("ssa-practice-controller-system")
		fieldMgr         = "ssapractice-fieldmanager"
		labels           = map[string]string{"apps": "ssapractice-nginx"}
		log              = r.Log.WithValues("ssapractice", req.NamespacedName)
		podTemplate      *corev1apply.PodTemplateSpecApplyConfiguration
		ssapractice      ssapracticev1.SSAPractice
	)

	err := r.Get(ctx, req.NamespacedName, &ssapractice)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "unable to fetch CR SSAPractice")
		return ctrl.Result{}, err
	}

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

	applied, err := deploymentClient.Apply(ctx, deploymentApplyConfig, metav1.ApplyOptions{
		FieldManager: fieldMgr,
		Force:        true,
	})
	if err != nil {
		log.Error(err, "unable to apply")
		return ctrl.Result{}, err
	}

	log.Info(fmt.Sprintf("Applied: %s", applied.GetName()))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SSAPracticeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ssapracticev1.SSAPractice{}).
		Complete(r)
}

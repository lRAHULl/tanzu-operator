/*
Copyright 2021.

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

	// "github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	tanzuv1alpha1 "tanzu-operator/api/v1alpha1"
)

// SonobuoyReconciler reconciles a Sonobuoy object
type SonobuoyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tanzu.coda.global,resources=sonobuoy,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tanzu.coda.global,resources=sonobuoy/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tanzu.coda.global,resources=sonobuoy/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *SonobuoyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling Sonobuoy Operator")

	// your logic here
	var sonobuoyOperator tanzuv1alpha1.Sonobuoy
	err := r.Get(ctx, req.NamespacedName, &sonobuoyOperator)
	if err != nil {
		fmt.Println(err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	} else {
		fmt.Println("NOERROR - Got Operator context")
	}

	var sa corev1.ServiceAccount
	sa.Name = sonobuoyOperator.Name
	sa.Namespace = sonobuoyOperator.Spec.Namespace

	_, err2 := ctrl.CreateOrUpdate(ctx, r.Client, &sa, func() error {
		return ctrl.SetControllerReference(&sonobuoyOperator, &sa, r.Scheme)
	})

	if err2 != nil {
		fmt.Println(err2.Error())
		return ctrl.Result{}, nil
	} else {
		fmt.Println("NOERROR - serviceaccount creation successful")
	}

	var clusterRole rbacv1.ClusterRole
	clusterRole.Name = sonobuoyOperator.Name
	clusterRole.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"namespaces",
				"services",
				"serviceaccounts",
				"configmaps",
			},
			Verbs: []string{
				"create",
				"delete",
				"get",
				"list",
				"update",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"clusterroles",
				"clusterrolebindings",
			},
			Verbs: []string{
				"create",
				"delete",
				"get",
				"list",
				"update",
				"watch",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"*",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
	}

	_, err3 := ctrl.CreateOrUpdate(ctx, r.Client, &clusterRole, func() error {
		return ctrl.SetControllerReference(&sonobuoyOperator, &clusterRole, r.Scheme)
	})

	if err3 != nil {
		fmt.Println(err3.Error())
		return ctrl.Result{}, nil
	} else {
		fmt.Println("NOERROR - Clusterrole creation successful")
	}

	var clusterRoleBinding rbacv1.ClusterRoleBinding
	clusterRoleBinding.Name = sonobuoyOperator.Name
	clusterRoleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      sa.Name,
			Namespace: sa.Namespace,
		},
	}
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRole.Name,
	}

	_, err4 := ctrl.CreateOrUpdate(ctx, r.Client, &clusterRoleBinding, func() error {
		return ctrl.SetControllerReference(&sonobuoyOperator, &clusterRoleBinding, r.Scheme)
	})

	if err4 != nil {
		fmt.Println(err4.Error())
		return ctrl.Result{}, nil
	} else {
		fmt.Println("NOERROR - Clusterrolebinding creation successful")
	}

	sonobuoyOptions := ""
	for key, val := range sonobuoyOperator.Spec.SonobuoyOptions {
		s := fmt.Sprintf(" %s %s ", key, val)
		sonobuoyOptions += s
	}
	var pod corev1.Pod
	pod.Name = sonobuoyOperator.Name
	pod.Namespace = sonobuoyOperator.Spec.Namespace
	pod.Spec.ServiceAccountName = sa.Name
	pod.Spec.Containers = []corev1.Container{
		{
			Name:      sonobuoyOperator.Name,
			Image:     sonobuoyOperator.Spec.Image,
			Resources: *sonobuoyOperator.Spec.Resources.DeepCopy(),
			Command: []string{
				"sh",
				"-c",
				"sonobuoy " + sonobuoyOperator.Spec.SonobuoyMode + sonobuoyOptions + " && while true; do sleep 10000; done ",
			},
		},
	}

	_, err1 := ctrl.CreateOrUpdate(ctx, r.Client, &pod, func() error {
		return ctrl.SetControllerReference(&sonobuoyOperator, &pod, r.Scheme)
	})

	if err1 != nil {
		fmt.Println(err1.Error())
		return ctrl.Result{}, nil
	} else {
		fmt.Println("NOERROR - Pod creation successful")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SonobuoyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tanzuv1alpha1.Sonobuoy{}).
		Complete(r)
}

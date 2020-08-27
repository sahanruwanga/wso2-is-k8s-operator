/*


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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wso2v1 "github.com/tsuresh/wso2-is-k8s-operator/api/v1"
)

// Wso2IsReconciler reconciles a Wso2Is object
type Wso2IsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=wso2.wso2.com,resources=wso2is,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wso2.wso2.com,resources=wso2is/status,verbs=get;update;patch

func (r *Wso2IsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("wso2is", req.NamespacedName)

	// your logic here
	// Fetch the WSO2IS instance
	instance := wso2v1.Wso2Is{}
	instance.Namespace = instance.Spec.Namespace

	// Check if WSO2 custom resource is present
	err := r.Get(ctx, req.NamespacedName, &instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("WSO2IS resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get WSO2IS Instance")
		return ctrl.Result{}, err
	}

	// Add new service account if not present
	svcFound := &corev1.ServiceAccount{}
	err = r.Get(ctx, types.NamespacedName{Name: "wso2svc-account", Namespace: instance.Namespace}, svcFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		svc := r.addServiceAccount(instance)
		log.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", svc.Namespace, "ServiceAccount.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new ServiceAccount", "ServiceAccount.Namespace", svc.Namespace, "ServiceAccount.Name", svc.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Successfully created new ServiceAccount", "ServiceAccount.Namespace", svc.Namespace, "ServiceAccount.Name", svc.Name)
		}
		// ServiceAccount created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ServiceAccount")
		return ctrl.Result{}, err
	}

	// Add new config map if not present
	confMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: "identity-server-conf", Namespace: instance.Namespace}, confMap)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		svc := r.addConfigMap(instance)
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", svc.Namespace, "ConfigMap.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", svc.Namespace, "ConfigMap.Name", svc.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Successfully created new ConfigMap", "ConfigMap.Namespace", svc.Namespace, "ConfigMap.Name", svc.Name)
		}
		// ServiceAccount created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// Add new service if not present
	serviceFound := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: "wso2is-service", Namespace: instance.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		svc := r.addNewService(instance)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Successfully created new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		}
		// ServiceAccount created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Add Ingress if not present
	ingressFound := v1beta1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: "wso2is-ingress", Namespace: instance.Namespace}, &ingressFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Ingress
		svc := r.addNewIngress(instance)
		log.Info("Creating new Ingress", "Ingress.Namespace", svc.Namespace, "Ingress.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", svc.Namespace, "Ingress.Name", svc.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Successfully created new Ingress", "Ingress.Namespace", svc.Namespace, "Ingress.Name", svc.Name)
		}
		// Ingress created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Ingress")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForWso2Is(instance)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		} else {
			log.Info("Successfully added new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := instance.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the IS status with the pod names
	// List the pods for this IS's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelsForWso2IS(instance.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "WSO2IS.Namespace", instance.Namespace, "WSO2IS.Name", instance.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.Status().Update(ctx, &instance)
		if err != nil {
			log.Error(err, "Failed to update WSO2IS status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// labelsForWso2IS returns the labels for selecting the resources
// belonging to the given WSO2IS CR name.
func labelsForWso2IS(name string) map[string]string {
	return map[string]string{"app": "wso2is", "wso2is_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// addServiceAccount adds a new ServiceAccount
func (r *Wso2IsReconciler) addServiceAccount(m wso2v1.Wso2Is) *corev1.ServiceAccount {
	svc := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wso2svc-account",
			Namespace: m.Namespace,
		},
	}
	ctrl.SetControllerReference(&m, svc, r.Scheme)
	return svc
}

// addConfigMap adds a new ConfigMap
func (r *Wso2IsReconciler) addConfigMap(m wso2v1.Wso2Is) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "identity-server-conf",
			Namespace: m.Namespace,
		},
		Data: map[string]string{
			"deployment.toml": "|-\n    [server]\n    hostname = \"$env{HOST_NAME}\"\n    node_ip = \"$env{NODE_IP}\"\n    # base_path = \"https://$ref{server.hostname}:${carbon.management.port}\"\n    [super_admin]\n    username = \"admin\"\n    password = \"admin\"\n    create_admin_account = true\n    [user_store]\n    type = \"read_write_ldap_unique_id\"\n    connection_url = \"ldap://localhost:${Ports.EmbeddedLDAP.LDAPServerPort}\"\n    connection_name = \"uid=admin,ou=system\"\n    connection_password = \"admin\"\n    base_dn = \"dc=wso2,dc=org\"      #refers the base dn on which the user and group search bases will be generated\n    [database.identity_db]\n    type = \"h2\"\n    url = \"jdbc:h2:./repository/database/WSO2IDENTITY_DB;DB_CLOSE_ON_EXIT=FALSE;LOCK_TIMEOUT=60000\"\n    username = \"wso2carbon\"\n    password = \"wso2carbon\"\n    [database.shared_db]\n    type = \"h2\"\n    url = \"jdbc:h2:./repository/database/WSO2SHARED_DB;DB_CLOSE_ON_EXIT=FALSE;LOCK_TIMEOUT=60000\"\n    username = \"wso2carbon\"\n    password = \"wso2carbon\"\n    [keystore.primary]\n    file_name = \"wso2carbon.jks\"\n    password = \"wso2carbon\"",
		},
	}
	ctrl.SetControllerReference(&m, configMap, r.Scheme)
	return configMap
}

// addServiceAccount adds a new ServiceAccount
func (r *Wso2IsReconciler) addNewIngress(m wso2v1.Wso2Is) *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wso2is-ingress",
			Namespace: m.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                     "nginx",
				"nginx.ingress.kubernetes.io/backend-protocol":    "HTTPS",
				"nginx.ingress.kubernetes.io/affinity":            "cookie",
				"nginx.ingress.kubernetes.io/session-cookie-name": "route",
				"nginx.ingress.kubernetes.io/session-cookie-hash": "sha1",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts: []string{"wso2is"},
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: "wso2is",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{{
								Path: "/",
								Backend: v1beta1.IngressBackend{
									ServiceName: "wso2is-service",
									ServicePort: intstr.IntOrString{
										IntVal: 9443,
									},
								},
							}},
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(&m, ingress, r.Scheme)
	return ingress
}

// addServiceAccount adds a new ServiceAccount
func (r *Wso2IsReconciler) addNewService(m wso2v1.Wso2Is) *corev1.Service {
	serviceType := corev1.ServiceTypeNodePort
	if m.Spec.ServiceType == "loadbalancer" {
		serviceType = corev1.ServiceTypeLoadBalancer
	} else if m.Spec.ServiceType == "clusterIP" {
		serviceType = corev1.ServiceTypeClusterIP
	} else if m.Spec.ServiceType == "externalName" {
		serviceType = corev1.ServiceTypeExternalName
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wso2is-service",
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "servlet-http",
				Protocol: "TCP",
				Port:     9763,
				TargetPort: intstr.IntOrString{
					IntVal: 9763,
				},
			}, {
				Name:     "servlet-https",
				Protocol: "TCP",
				Port:     9443,
				TargetPort: intstr.IntOrString{
					IntVal: 9443,
				},
			}},
			Selector: map[string]string{
				"deployment": m.Name,
				"app":        "wso2is",
				"monitoring": "jmx",
				"pod":        "wso2is-sample",
			},
			Type: serviceType,
		},
	}
	ctrl.SetControllerReference(&m, svc, r.Scheme)
	return svc
}

// New deployment for WSO2IS
func (r *Wso2IsReconciler) deploymentForWso2Is(m wso2v1.Wso2Is) *appsv1.Deployment {
	ls := labelsForWso2IS(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "wso2/wso2is",
						Name:  "wso2is",
						//Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9443,
							Name:          "wso2is",
						}},
					}},
				},
			},
		},
	}
	// Set WSO2IS instance as the owner and controller
	ctrl.SetControllerReference(&m, dep, r.Scheme)
	return dep
}

func (r *Wso2IsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wso2v1.Wso2Is{}).
		Complete(r)
}

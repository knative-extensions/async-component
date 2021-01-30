/*
Copyright 2020 The Knative Authors

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

package ingress

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/net-contour/pkg/reconciler/contour/config"
	fakenetworkingclient "knative.dev/networking/pkg/client/injection/client/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"

	ktesting "k8s.io/client-go/testing"
	ingressreconciler "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	. "knative.dev/async-component/pkg/reconciler/testing"
	network "knative.dev/networking/pkg"

	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	netv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	. "knative.dev/pkg/reconciler/testing"
)

type testConfigStore struct {
	config *config.Config
}

const (
	defaultNamespace       = "default"
	testingName            = "testing"
	testingAlwaysAsyncName = "testing-always"
	knativeTesting         = "knative-testing"
	exampleHost            = "example.com"
	testHost               = "test.com"
	serviceName            = "servicename"
)

var statusReady = v1alpha1.IngressStatus{
	PublicLoadBalancer: &v1alpha1.LoadBalancerStatus{
		Ingress: []v1alpha1.LoadBalancerIngressStatus{
			{DomainInternal: publicLBDomain},
		},
	},
	PrivateLoadBalancer: &v1alpha1.LoadBalancerStatus{
		Ingress: []v1alpha1.LoadBalancerIngressStatus{
			{DomainInternal: privateLBDomain},
		},
	},
	Status: duckv1.Status{
		Conditions: duckv1.Conditions{{
			Type:   v1alpha1.IngressConditionLoadBalancerReady,
			Status: corev1.ConditionTrue,
		}, {
			Type:   v1alpha1.IngressConditionNetworkConfigured,
			Status: corev1.ConditionTrue,
		}, {
			Type:   v1alpha1.IngressConditionReady,
			Status: corev1.ConditionTrue,
		}},
	},
}

var statusUnknown = v1alpha1.IngressStatus{
	Status: duckv1.Status{
		Conditions: duckv1.Conditions{{
			Type:   v1alpha1.IngressConditionLoadBalancerReady,
			Status: corev1.ConditionUnknown,
		}, {
			Type:   v1alpha1.IngressConditionNetworkConfigured,
			Status: corev1.ConditionUnknown,
		}, {
			Type:   v1alpha1.IngressConditionReady,
			Status: corev1.ConditionUnknown,
		}},
	},
}

var ingWithAsyncAnnotation = ingress(defaultNamespace, testingName, statusReady,
	withAnnotations(map[string]string{
		networking.IngressClassAnnotationKey: asyncIngressClassName,
	}))
var ingAlwaysAsync = ingress(defaultNamespace, testingAlwaysAsyncName, statusReady,
	withAnnotations(map[string]string{
		networking.IngressClassAnnotationKey: asyncIngressClassName,
		AsyncModeAnnotationKey:               asyncAlwaysMode,
	}),
)
var ingSometimesAsync = ingress(defaultNamespace, testingName, statusReady,
	withAnnotations(map[string]string{
		networking.IngressClassAnnotationKey: asyncIngressClassName,
		AsyncModeAnnotationKey:               asyncConditionalMode,
	}),
)
var ingInvalidModeAnnotation = ingress(defaultNamespace, testingName, statusReady,
	withAnnotations(map[string]string{
		networking.IngressClassAnnotationKey: asyncIngressClassName,
		AsyncModeAnnotationKey:               "invalid.mode.annotation.value",
	}),
)

var createdIng = expectedCreatedIng(defaultNamespace, testingName, statusUnknown, false)
var createdIngWithAsyncAlways = expectedCreatedIng(defaultNamespace, testingAlwaysAsyncName, statusUnknown, true)

func TestReconcile(t *testing.T) {
	createdIng.Status.InitializeConditions()
	changedService := service(defaultNamespace, testingName)
	changedService.Spec.ExternalName = "changed"
	table := TableTest{
		{
			Name: "skip ingress not matching class key",
			Objects: []runtime.Object{
				ingress("testing", "testing", statusReady, withAnnotations(
					map[string]string{networking.IngressClassAnnotationKey: "fake-class-annotation"})),
			},
		},
		{
			Name: "create new ingress with async annotation",
			Key:  "default/testing",
			Objects: []runtime.Object{
				ingWithAsyncAnnotation,
			},
			WantCreates: []runtime.Object{
				createdIng,
				service(defaultNamespace, testingName),
			},
		},
		{
			Name: "test service update",
			Key:  "default/testing",
			Objects: []runtime.Object{
				ingWithAsyncAnnotation,
				changedService,
			},
			WantCreates: []runtime.Object{
				createdIng,
			},
			WantUpdates: []ktesting.UpdateActionImpl{{
				Object: service(defaultNamespace, testingName),
			}},
		},
		{
			Name: "create new ingress with async annotation and sometimes mode value",
			Key:  "default/testing",
			Objects: []runtime.Object{
				ingSometimesAsync,
			},
			WantCreates: []runtime.Object{
				createdIng,
				service(defaultNamespace, testingName),
			},
		},
		{
			Name: "create new ingress with async annotation and always mode value",
			Key:  "default/testing-always",
			Objects: []runtime.Object{
				ingAlwaysAsync,
			},
			WantCreates: []runtime.Object{
				createdIngWithAsyncAlways,
				service(defaultNamespace, testingAlwaysAsyncName),
			},
		},
		{
			Name: "create new ingress with async annotation and invalid mode value",
			Key:  "default/testing",
			Objects: []runtime.Object{
				ingInvalidModeAnnotation,
			},
			WantErr: true,
			WantEvents: []string{
				Eventf(corev1.EventTypeWarning, "InternalError", "Invalid value for key async.knative.dev/mode: "),
			},
		},
	}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			netclient:     fakenetworkingclient.Get(ctx),
			ingressLister: listers.GetIngressLister(),
			serviceLister: listers.GetK8sServiceLister(),
			kubeclient:    fakekubeclient.Get(ctx),
		}
		return ingressreconciler.NewReconciler(ctx, logging.FromContext(ctx), fakenetworkingclient.Get(ctx),
			listers.GetIngressLister(), controller.GetEventRecorder(ctx), r, asyncIngressClassName, controller.Options{})
	}))
}

type ingressCreationOption func(ing *v1alpha1.Ingress)

func ingress(namespace, name string, status v1alpha1.IngressStatus, opt ...ingressCreationOption) *v1alpha1.Ingress {
	ing := &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: netv1alpha1.IngressSpec{
			Rules: []netv1alpha1.IngressRule{{
				Hosts:      []string{exampleHost},
				Visibility: netv1alpha1.IngressVisibilityExternalIP,
				HTTP: &netv1alpha1.HTTPIngressRuleValue{
					Paths: []netv1alpha1.HTTPIngressPath{{
						Splits: []netv1alpha1.IngressBackendSplit{{
							Percent: 100,
							AppendHeaders: map[string]string{
								network.OriginalHostHeader: testHost,
							},
							IngressBackend: netv1alpha1.IngressBackend{
								ServiceName:      serviceName,
								ServiceNamespace: namespace,
								ServicePort:      intstr.FromInt(80),
							},
						}},
					}},
				},
			}},
		},
		Status: status,
	}
	for _, o := range opt {
		o(ing)
	}
	return ing
}

func withAnnotations(ans map[string]string) ingressCreationOption {
	return func(ing *v1alpha1.Ingress) {
		ing.Annotations = ans
	}
}

func expectedCreatedIng(namespace, name string, status v1alpha1.IngressStatus, isAlwaysAsync bool) *v1alpha1.Ingress {
	thePaths := []netv1alpha1.HTTPIngressPath{}
	if isAlwaysAsync {
		thePaths = append([]netv1alpha1.HTTPIngressPath{{
			Headers: map[string]v1alpha1.HeaderMatch{preferHeaderField: {Exact: preferSyncValue}},
			Splits: []netv1alpha1.IngressBackendSplit{{
				IngressBackend: v1alpha1.IngressBackend{
					ServiceName:      serviceName,
					ServiceNamespace: namespace,
					ServicePort:      intstr.FromInt(80),
				},
				Percent:       int(100),
				AppendHeaders: map[string]string{"K-Original-Host": testHost},
			}},
		},
			{
				RewriteHost: getClusterLocalDomain(producerServiceName, knativeTesting),
				Splits: []netv1alpha1.IngressBackendSplit{{
					Percent: 100,
					IngressBackend: netv1alpha1.IngressBackend{
						ServiceNamespace: namespace,
						ServiceName:      name + asyncSuffix,
						ServicePort:      intstr.FromInt(80),
					},
				},
				},
				AppendHeaders: map[string]string{asyncOriginalHostHeader: getClusterLocalDomain(name, defaultNamespace)},
			},
		})
	} else {
		thePaths = append([]netv1alpha1.HTTPIngressPath{{
			RewriteHost: getClusterLocalDomain(producerServiceName, knativeTesting),
			Headers:     map[string]v1alpha1.HeaderMatch{preferHeaderField: {Exact: preferAsyncValue}},
			Splits: []netv1alpha1.IngressBackendSplit{
				{
					IngressBackend: v1alpha1.IngressBackend{
						ServiceName:      name + asyncSuffix,
						ServiceNamespace: namespace,
						ServicePort:      intstr.FromInt(80),
					},
					Percent: int(100),
				}},
			AppendHeaders: map[string]string{
				asyncOriginalHostHeader: getClusterLocalDomain(name, defaultNamespace),
			},
		},
			{Splits: []netv1alpha1.IngressBackendSplit{
				{
					Percent: 100,
					AppendHeaders: map[string]string{
						network.OriginalHostHeader: testHost,
					},
					IngressBackend: netv1alpha1.IngressBackend{
						ServiceNamespace: namespace,
						ServiceName:      serviceName,
						ServicePort:      intstr.FromInt(80),
					},
				},
			}},
		})
	}
	return &netv1alpha1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + newSuffix,
			Namespace:   namespace,
			Annotations: map[string]string{networking.IngressClassAnnotationKey: network.IstioIngressClassName},
		},
		Spec: netv1alpha1.IngressSpec{
			Rules: []netv1alpha1.IngressRule{{
				Hosts:      []string{exampleHost},
				Visibility: netv1alpha1.IngressVisibilityExternalIP,
				HTTP: &netv1alpha1.HTTPIngressRuleValue{
					Paths: thePaths,
				},
			}},
		},
		Status: status,
	}
}

func service(namespace, name string) *corev1.Service {
	selector := make(map[string]string)
	selector["app"] = producerServiceName
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + asyncSuffix,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: getClusterLocalDomain(producerServiceName, knativeTesting),
			Ports: []corev1.ServicePort{{
				Name:       networking.ServicePortName(networking.ProtocolHTTP1),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(networking.ServicePort(networking.ProtocolHTTP1)),
				TargetPort: intstr.FromInt(80),
			}},
			Selector:        selector,
			SessionAffinity: "None",
		},
	}
	return svc
}
